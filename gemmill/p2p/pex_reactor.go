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

package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"

	"go.uber.org/zap"
)

var pexErrInvalidMessage = errors.New("Invalid PEX message")

const (
	PexChannel               = byte(0x00)
	ensurePeersPeriodSeconds = 30
	minNumOutboundPeers      = 10
	maxPexMessageSize        = 1048576 // 1MB
)

/*
PEXReactor handles PEX (peer exchange) and ensures that an
adequate number of peers are connected to the switch.
*/
type PEXReactor struct {
	BaseReactor

	book *AddrBook
}

func NewPEXReactor(book *AddrBook) *PEXReactor {
	pexR := &PEXReactor{
		book: book,
	}
	pexR.BaseReactor = *NewBaseReactor("PEXReactor", pexR)
	return pexR
}

func (pexR *PEXReactor) OnStart() error {
	pexR.BaseReactor.OnStart()
	go pexR.ensurePeersRoutine()
	return nil
}

func (pexR *PEXReactor) OnStop() {
	pexR.BaseReactor.OnStop()
}

// Implements Reactor
func (pexR *PEXReactor) GetChannels() []*ChannelDescriptor {
	return []*ChannelDescriptor{
		&ChannelDescriptor{
			ID:                PexChannel,
			Priority:          1,
			SendQueueCapacity: 10,
		},
	}
}

// Implements Reactor
func (pexR *PEXReactor) AddPeer(peer *Peer) {
	// Add the peer to the address book
	netAddr, _ := NewNetAddressString(peer.ListenAddr)
	if peer.IsOutbound() {
		if pexR.book.NeedMoreAddrs() {
			pexR.RequestPEX(peer)
		}
	} else {
		// For inbound connections, the peer is its own source
		// (For outbound peers, the address is already in the books)
		pexR.book.AddAddress(netAddr, netAddr)
	}
}

// Implements Reactor
func (pexR *PEXReactor) RemovePeer(peer *Peer, reason interface{}) {
	// TODO
}

// Implements Reactor
// Handles incoming PEX messages.
func (pexR *PEXReactor) Receive(chID byte, src *Peer, msgBytes []byte) {
	start := time.Now()
	defer func() {
		src.AuditLog(chID, msgBytes, start, pexR.String())
	}()
	// decode message
	_, msg, err := DecodeMessage(msgBytes)
	if err != nil {
		log.Warn("Error decoding message", zap.String("error", err.Error()))
		return
	}
	log.Infow("Received message", "msg", msg)

	switch msg := msg.(type) {
	case *pexRequestMessage:
		// src requested some peers.
		// TODO: prevent abuse.
		pexR.SendAddrs(src, pexR.book.GetSelection())
	case *pexAddrsMessage:
		// We received some peer addresses from src.
		// TODO: prevent abuse.
		// (We don't want to get spammed with bad peers)
		srcAddr := src.Connection().RemoteAddress
		for _, addr := range msg.Addrs {
			pexR.book.AddAddress(addr, srcAddr)
		}
	default:
		log.Warnf("Unknown message type %T", msg)
	}

}

// Asks peer for more addresses.
func (pexR *PEXReactor) RequestPEX(peer *Peer) {
	peer.Send(PexChannel, struct{ PexMessage }{&pexRequestMessage{}})
}

func (pexR *PEXReactor) SendAddrs(peer *Peer, addrs []*NetAddress) {
	peer.Send(PexChannel, struct{ PexMessage }{&pexAddrsMessage{Addrs: addrs}})
}

// Ensures that sufficient peers are connected. (continuous)
func (pexR *PEXReactor) ensurePeersRoutine() {
	// Randomize when routine starts
	time.Sleep(time.Duration(rand.Int63n(500*ensurePeersPeriodSeconds)) * time.Millisecond)

	// fire once immediately.
	pexR.ensurePeers()
	// fire periodically
	timer := gcmn.NewRepeatTimer("pex", ensurePeersPeriodSeconds*time.Second)
FOR_LOOP:
	for {
		select {
		case <-timer.Ch:
			pexR.ensurePeers()
		case <-pexR.Quit:
			break FOR_LOOP
		}
	}

	// Cleanup
	timer.Stop()
}

// Ensures that sufficient peers are connected. (once)
func (pexR *PEXReactor) ensurePeers() {
	numOutPeers, _, numDialing := pexR.Switch.NumPeers()
	numToDial := minNumOutboundPeers - (numOutPeers + numDialing)
	log.Debug("Ensure peers", zap.Int("numOutPeers", numOutPeers), zap.Int("numDialing", numDialing), zap.Int("numToDial", numToDial))
	if numToDial <= 0 {
		return
	}
	toDial := make(map[string]*NetAddress)

	myAddr := pexR.Switch.nodeInfo.ListenAddr

	// Try to pick numToDial addresses to dial.
	for i := 0; i < numToDial; i++ {
		// The purpose of newBias is to first prioritize old (more vetted) peers
		// when we have few connections, but to allow for new (less vetted) peers
		// if we already have many connections. This algorithm isn't perfect, but
		// it somewhat ensures that we prioritize connecting to more-vetted
		// peers.

		newBias := gcmn.MinInt(numOutPeers, 8)*10 + 10
		var picked *NetAddress
		// Try to fetch a new peer 3 times.
		// This caps the maximum number of tries to 3 * numToDial.
		for j := 0; j < 3; j++ {
			try := pexR.book.PickAddress(newBias)
			if try == nil {
				break
			}
			tryAddr := try.String()
			_, alreadySelected := toDial[tryAddr]
			alreadyDialing := pexR.Switch.IsDialing(try)
			alreadyConnected := pexR.Switch.Peers().Has(tryAddr)
			if myAddr == tryAddr || alreadySelected || alreadyDialing || alreadyConnected {
				// pexR.Logger.Info("Cannot dial address", "addr", try,
				//      "alreadySelected", alreadySelected,
				//      "alreadyDialing", alreadyDialing,
				//  "alreadyConnected", alreadyConnected)

				continue
			} else {
				log.Debug("Will dial address", zap.Stringer("addr", try))
				picked = try
				break
			}
		}
		if picked == nil {
			continue
		}
		toDial[picked.String()] = picked
	}

	// Dial picked addresses
	for _, item := range toDial {
		go func(picked *NetAddress) {
			_, err := pexR.Switch.DialPeerWithAddress(picked)
			if err != nil {
				pexR.book.MarkAttempt(picked)
			}
		}(item)
	}

	// If we need more addresses, pick a random peer and ask for more.
	if pexR.book.NeedMoreAddrs() {
		if peers := pexR.Switch.Peers().List(); len(peers) > 0 {
			i := rand.Int() % len(peers)
			peer := peers[i]
			log.Debug("No addresses to dial. Sending pexRequest to random peer", zap.Stringer("peer", peer))
			pexR.RequestPEX(peer)
		}
	}
}

//-----------------------------------------------------------------------------
// Messages

const (
	msgTypeRequest = byte(0x01)
	msgTypeAddrs   = byte(0x02)
)

type PexMessage interface{}

var _ = wire.RegisterInterface(
	struct{ PexMessage }{},
	wire.ConcreteType{&pexRequestMessage{}, msgTypeRequest},
	wire.ConcreteType{&pexAddrsMessage{}, msgTypeAddrs},
)

func DecodeMessage(bz []byte) (msgType byte, msg PexMessage, err error) {
	msgType = bz[0]
	n := new(int)
	r := bytes.NewReader(bz)
	msg = wire.ReadBinary(struct{ PexMessage }{}, r, maxPexMessageSize, n, &err).(struct{ PexMessage }).PexMessage
	return
}

/*
A pexRequestMessage requests additional peer addresses.
*/
type pexRequestMessage struct {
}

func (m *pexRequestMessage) String() string {
	return "[pexRequest]"
}

/*
A message with announced peer addresses.
*/
type pexAddrsMessage struct {
	Addrs []*NetAddress
}

func (m *pexAddrsMessage) String() string {
	return fmt.Sprintf("[pexAddrs %v]", m.Addrs)
}
