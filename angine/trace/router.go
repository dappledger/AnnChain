// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	trcpb "github.com/dappledger/AnnChain/angine/protos/trace"
	sm "github.com/dappledger/AnnChain/angine/state"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
)

// Router keeps tracks of the messages flow around in the TraceReactor.
// Messages are wrapped in the "TraceMessage" envelope universally. And messages are identified by Hash uniquely in the Router.
// Different purposes and related handlers are separated by ChannelID.
type Router struct {
	cmn.BaseService

	reactor       *Reactor
	config        *viper.Viper
	logger        *zap.Logger
	mtx           sync.Mutex
	state         *sm.State
	privValidator *agtypes.PrivValidator
	address       []byte

	// TODO: replace with some LRU data structure to keep consumption manageable
	// requestRouteTable map[string][]byte
	requestRouteTable *cache.Cache
	responseChannels  map[string]chan []byte
	channelHandlers   map[byte]ChannelResponseHandler
}

// ChannelResponseHandler defines the signature of the handler for a channel.
// Very cheap design, just []byte in & out.
type ChannelResponseHandler func([]byte) []byte

// NewRouter defines the constructor of Router, initialization.
func NewRouter(logger *zap.Logger, config *viper.Viper, state *sm.State, priv *agtypes.PrivValidator) *Router {
	msgTTL := time.Duration(config.GetInt("tracerouter_msg_ttl"))
	router := &Router{
		logger:            logger,
		config:            config,
		state:             state,
		privValidator:     priv,
		address:           priv.GetAddress(),
		requestRouteTable: cache.New(msgTTL*time.Second, 100*msgTTL*time.Second),
		responseChannels:  make(map[string]chan []byte),
		channelHandlers:   make(map[byte]ChannelResponseHandler),
	}
	router.BaseService = *cmn.NewBaseService(logger, "TraceRouter", router)
	return router
}

// SetReactor binds with a trace reactor.
func (rt *Router) SetReactor(r *Reactor) {
	rt.reactor = r
}

// Requesthash hashes a message in []byte as the identification of this message.
func (rt *Router) Requesthash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Broadcast is the entry for start a traced communication.
// thread-safe
func (rt *Router) Broadcast(ctx context.Context, data []byte, voteCh chan []byte) error {
	msg := &trcpb.TraceRequest{Data: data}
	hash := HashMessage(msg)

	rt.mtx.Lock()
	if _, ok := rt.responseChannels[string(hash)]; ok {
		rt.mtx.Unlock()
		return errors.Errorf("TraceState: duplicate trace data [%X]", hash)
	}
	rt.responseChannels[string(hash)] = voteCh
	rt.mtx.Unlock()

	rt.reactor.Switch.BroadcastBytes(SpecialOPChannel, trcpb.MarshalDataToTrcMsg(msg))
	return nil
}

// RegisterHandler binds a channel with the specified handler.
func (rt *Router) RegisterHandler(chID byte, fn ChannelResponseHandler) error {
	for _, desciption := range rt.reactor.GetChannels() {
		if desciption.ID == chID {
			rt.channelHandlers[chID] = fn
			return nil
		}
	}
	return errors.Errorf("no such channel: %v", chID)
}

// SetPrivValidator
func (rt *Router) SetPrivValidator(priv *agtypes.PrivValidator) {
	rt.privValidator = priv
	rt.address = priv.GetAddress()
}

// TraceRequest is a thread-safe wrapper of the processing of request.
func (rt *Router) TraceRequest(requestHash []byte, fromPeer []byte, data []byte, onlyValidator bool) (broadcast bool, err error) {
	rt.mtx.Lock()
	defer rt.mtx.Unlock()
	return rt.request(requestHash, fromPeer, data, onlyValidator)
}

// request checks if the node has got the message already, if so, throws this message.
// and if it's a new message, then the node keeps a new track of the message, and broadcasts the message to its peers to move forward the message.
// after that, the node needs to check if it is a validator.
// if positive:
//   1. find the handler bound with the channel
//   2. handle the message
//   3. consutruct a "trace response" and send it back to where the message comes from, no broadcasts
func (rt *Router) request(requestHash []byte, fromPeer []byte, data []byte, onlyValidator bool) (broadcast bool, err error) {
	if _, ok := rt.requestRouteTable.Get(string(requestHash)); ok {
		// ignore duplicated msg
		return false, nil
	}
	rt.requestRouteTable.Set(string(requestHash), fromPeer, cache.DefaultExpiration)
	if onlyValidator && !rt.isValidator() {
		// just let the reator keep broadcasting the message
		return true, nil
	}

	// sign the change validator msg
	resp := &trcpb.TraceResponse{
		RequestHash: requestHash,
		Resp:        rt.channelHandlers[SpecialOPChannel](data),
	}

	// peer is keyed by its PublicKey in the form of uppercase hex string
	rt.reactor.Switch.Peers().Get(fmt.Sprintf("%X", fromPeer)).SendBytes(SpecialOPChannel, trcpb.MarshalDataToTrcMsg(resp))
	return true, nil
}

// TraceRespond is a thread-safe wrapper of handling response
func (rt *Router) TraceRespond(requestHash []byte, data []byte) ([]byte, error) {
	rt.mtx.Lock()
	defer rt.mtx.Unlock()

	return rt.respond(requestHash, data)
}

// responed takes care of 2 things.
// 1. if the node is the one who starts the request, send the data into the corresponding channel
// 2. finds the peer who send me the request and gives back this reponse for retrospect
func (rt *Router) respond(requestHash []byte, data []byte) ([]byte, error) {
	if ch, ok := rt.responseChannels[string(requestHash)]; ok {
		ch <- data
		return nil, nil
	}
	peer, ok := rt.requestRouteTable.Get(string(requestHash))
	if !ok {
		return nil, nil
	}

	return peer.([]byte), nil
}

func (rt *Router) isValidator() bool {
	return rt.state.Validators.HasAddress(rt.address)
}
