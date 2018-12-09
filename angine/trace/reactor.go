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
	"crypto/sha256"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	trcpb "github.com/dappledger/AnnChain/angine/protos/trace"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
)

const (
	// SpecialOPChannel is for special operation only
	SpecialOPChannel = byte(0x50)

	maxTraceMessageSize = 1048576
)

type channelAttribute struct {
	ValidatorOnly bool
}

type Reactor struct {
	// every reactor implements p2p.BaseReactor
	p2p.BaseReactor

	config *viper.Viper
	logger *zap.Logger

	evsw agtypes.EventSwitch

	// router keeps the state of all tracks
	router *Router

	channelAttributes map[byte]channelAttribute
}

func newChannelAttributes(r *Reactor) map[byte]channelAttribute {
	ret := make(map[byte]channelAttribute)
	ret[SpecialOPChannel] = channelAttribute{
		ValidatorOnly: true,
	}
	return ret
}

func NewTraceReactor(logger *zap.Logger, config *viper.Viper, r *Router) *Reactor {
	tr := &Reactor{
		logger:            logger,
		config:            config,
		router:            r,
		channelAttributes: make(map[byte]channelAttribute),
	}
	tr.BaseReactor = *p2p.NewBaseReactor(logger, "TraceReactor", tr)
	tr.channelAttributes = newChannelAttributes(tr)
	return tr
}

func (tr *Reactor) OnStart() error {
	return tr.BaseReactor.OnStart()
}

func (tr *Reactor) OnStop() {
	tr.BaseReactor.OnStop()
}

// GetChannels returns channel descriptor for all supported channels.
// maybe more in the furture
func (tr *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID:                SpecialOPChannel,
			Priority:          1,
			SendQueueCapacity: 100,
		},
	}
}

// AddPeer has nothing to do
func (tr *Reactor) AddPeer(peer *p2p.Peer) {}

// RemovePeer has nothing to do
func (tr *Reactor) RemovePeer(peer *p2p.Peer, reason interface{}) {}

// Receive is the main entrance of handling data flow
func (tr *Reactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	msg, err := trcpb.UnmarshalTrcMsg(msgBytes)
	if err != nil {
		tr.logger.Warn("error decoding message", zap.Error(err))
		return
	}

	// different channel different strategy,
	// mainly focus on how to broadcast the message
	switch chID {
	case SpecialOPChannel:
		switch msg := msg.(type) {
		case *trcpb.TraceRequest:
			broadcast, err := tr.router.TraceRequest(HashMessage(msg), src.PubKey[:], msg.Data, tr.channelAttributes[SpecialOPChannel].ValidatorOnly)
			if err != nil {
				// process error
				return
			}
			if broadcast {
				for _, p := range tr.Switch.Peers().List() {
					if !p.Equals(src) {
						p.SendBytes(SpecialOPChannel, trcpb.MarshalDataToTrcMsg(msg))
					}
				}
			}
		case *trcpb.TraceResponse:
			pk, err := tr.router.TraceRespond(msg.RequestHash, msg.Resp)
			if err != nil {
				// process error
				return
			}
			if pk != nil {
				peer := tr.Switch.Peers().Get(fmt.Sprintf("%X", pk))
				peer.SendBytes(SpecialOPChannel, trcpb.MarshalDataToTrcMsg(msg))
			}
		}
	default:
		// by default, we have nothing to do
	}
}

func (tr *Reactor) SetEventSwitch(evsw agtypes.EventSwitch) {
	tr.evsw = evsw
}

func HashMessage(msg trcpb.TraceMsgItfc) []byte {
	bys := trcpb.MarshalDataToTrcMsg(msg)
	hasher := sha256.New()
	hasher.Write(bys)
	return hasher.Sum(nil)
}
