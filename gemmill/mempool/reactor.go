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

package mempool

import (
	"bytes"
	"fmt"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/modules/go-clist"
	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/types"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	MempoolChannel = byte(0x30)

	maxMempoolMessageSize      = 1048576 // 1MB TODO make it configurable
	peerCatchupSleepIntervalMS = 100     // If peer is behind, sleep this amount
)

// MempoolReactor handles mempool tx broadcasting amongst peers.
type MempoolReactor struct {
	p2p.BaseReactor
	config  *viper.Viper
	Mempool types.TxPool
	evsw    types.EventSwitch
}

func NewTxReactor(conf *viper.Viper, mempool types.TxPool) *MempoolReactor {
	memR := &MempoolReactor{
		config:  conf,
		Mempool: mempool,
	}
	memR.BaseReactor = *p2p.NewBaseReactor("MempoolReactor", memR)
	return memR
}

// Implements Reactor
func (memR *MempoolReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		&p2p.ChannelDescriptor{
			ID: MempoolChannel,
			//			Priority: 5,
			Priority: 1,
		},
	}
}

// Implements Reactor
func (memR *MempoolReactor) AddPeer(peer *p2p.Peer) {
	go memR.broadcastTxRoutine(peer)
}

// Implements Reactor
func (memR *MempoolReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
	// broadcast routine checks if peer is gone and returns
}

// Implements Reactor
func (memR *MempoolReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	_, msg, err := DecodeMessage(msgBytes)
	if err != nil {
		log.Warn("Error decoding message", zap.String("error", err.Error()))
		return
	}
	//log.Debugw("Receive", "src", src, "chId", chID, "msg", msg)

	switch msg := msg.(type) {
	case *TxMessage:
		if err := memR.Mempool.ReceiveTx(msg.Tx); err != nil {
			// Bad, seen, or conflicting tx.
			// log.Debug("Could not add tx", zap.ByteString("tx", msg.Tx))
			return
		}
		// log.Debug("Added valid tx", zap.ByteString("tx", msg.Tx))
		// broadcasting happens from go routines per peer
	default:
		log.Info(fmt.Sprintf("Unknown message type %T", msg))
	}
}

// Just an alias for CheckTx since broadcasting happens in peer routines
//func (memR *MempoolReactor) BroadcastTx(tx types.Tx) error {
//	return memR.Mempool.CheckTx(tx)
//}

type PeerState interface {
	GetHeight() int64
}

type Peer interface {
	IsRunning() bool
	Send(byte, interface{}) bool
	Get(string) interface{}
}

// Send new mempool txs to peer.
// TODO: Handle mempool or reactor shutdown?
// As is this routine may block forever if no new txs come in.
func (memR *MempoolReactor) broadcastTxRoutine(peer Peer) {
	if !memR.config.GetBool("mempool_broadcast") {
		return
	}

	var next *clist.CElement
	for {
		if !memR.IsRunning() || !peer.IsRunning() {
			return // Quit!
		}
		if next == nil {
			// This happens because the CElement we were looking at got
			// garbage collected (removed).  That is, .NextWait() returned nil.
			// Go ahead and start from the beginning.
			next = memR.Mempool.TxsFrontWait() // Wait until a tx is available
		}
		memTx := next.Value.(*types.TxInPool)
		// make sure the peer is up to date
		height := memTx.GetHeight()
		if peerState := peer.Get(types.PeerStateKey); peerState != nil {
			pState := peerState.(PeerState)
			if pState.GetHeight() < height-1 { // Allow for a lag of 1 block
				time.Sleep(peerCatchupSleepIntervalMS * time.Millisecond)
				continue
			}
		}
		// send memTx
		msg := &TxMessage{Tx: memTx.Tx}
		success := peer.Send(MempoolChannel, struct{ MempoolMessage }{msg})
		if !success {
			time.Sleep(peerCatchupSleepIntervalMS * time.Millisecond)
			continue
		}

		next = next.NextWait()
		continue
	}
}

// implements events.Eventable
func (memR *MempoolReactor) SetEventSwitch(evsw types.EventSwitch) {
	memR.evsw = evsw
}

//-----------------------------------------------------------------------------
// Messages

const (
	msgTypeTx = byte(0x01)
)

type MempoolMessage interface{}

var _ = wire.RegisterInterface(
	struct{ MempoolMessage }{},
	wire.ConcreteType{&TxMessage{}, msgTypeTx},
)

func DecodeMessage(bz []byte) (msgType byte, msg MempoolMessage, err error) {
	msgType = bz[0]
	n := new(int)
	r := bytes.NewReader(bz)
	msg = wire.ReadBinary(struct{ MempoolMessage }{}, r, maxMempoolMessageSize, n, &err).(struct{ MempoolMessage }).MempoolMessage
	return
}

//-------------------------------------

type TxMessage struct {
	Tx types.Tx
}

func (m *TxMessage) String() string {
	return fmt.Sprintf("[TxMessage %v]", m.Tx)
}
