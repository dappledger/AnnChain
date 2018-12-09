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

package consensus

import (
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"

	csspb "github.com/dappledger/AnnChain/angine/protos/consensus"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	auto "github.com/dappledger/AnnChain/module/lib/go-autofile"
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

//--------------------------------------------------------
// types and functions for savings consensus messages

type walMessageJson struct {
	WalType byte   `json:"wal_type"`
	JsonBys []byte `json:"json_str"`
}

type StWALMessage struct {
	WALMessage
}

func (sw StWALMessage) MarshalJSON() ([]byte, error) {
	var wmj walMessageJson
	wmj.WalType = walMsgType(sw.WALMessage)
	var err error
	if wmj.JsonBys, err = sw.WALMessage.MarshalJSON(); err != nil {
		return nil, err
	}
	return json.Marshal(&wmj)
}

func (sw *StWALMessage) UnmarshalJSON(data []byte) error {
	var dec walMessageJson
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	switch dec.WalType {
	case WALMsgTypeRoundState:
		msg := &agtypes.EventDataRoundState{}
		err = json.Unmarshal(dec.JsonBys, msg)
		sw.WALMessage = msg
	case WALMsgTypeMsgInfo:
		msg := &msgInfo{}
		err = json.Unmarshal(dec.JsonBys, msg)
		sw.WALMessage = msg
	case WALMsgTypeTimeoutInfo:
		msg := &timeoutInfo{}
		err = json.Unmarshal(dec.JsonBys, msg)
		sw.WALMessage = msg
	default:
		return errors.New("wrong type of walmessage")
	}
	return err
}

type WALMessage interface {
	json.Marshaler
	json.Unmarshaler
}

type TimedWALMessage struct {
	Time time.Time    `json:"time"`
	Msg  StWALMessage `json:"msg"`
}

func GenTimedWALMessage(msg WALMessage) (retMsg TimedWALMessage) {
	retMsg.Time = time.Now()
	retMsg.Msg = StWALMessage{msg}
	return
}

func (tw *TimedWALMessage) GetMsg() WALMessage {
	return tw.Msg.WALMessage
}

func (tw *TimedWALMessage) UnmarshalJSON(data []byte) error {
	st := struct {
		Time time.Time    `json:"time"`
		Msg  StWALMessage `json:"msg"`
	}{}
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	tw.Time = st.Time
	tw.Msg = st.Msg
	return nil
}

const (
	WALMsgTypeRoundState  = byte(0x01)
	WALMsgTypeMsgInfo     = byte(0x02)
	WALMsgTypeTimeoutInfo = byte(0x03)
)

func walMsgType(wmsg WALMessage) byte {
	switch wmsg.(type) {
	case *agtypes.EventDataRoundState:
		return WALMsgTypeRoundState
	case *msgInfo:
		return WALMsgTypeMsgInfo
	case *timeoutInfo:
		return WALMsgTypeTimeoutInfo
	}
	return byte(0x00)
}

//--------------------------------------------------------
// Simple write-ahead logger

// Write ahead logger writes msgs to disk before they are processed.
// Can be used for crash-recovery and deterministic replay
// TODO: currently the wal is overwritten during replay catchup
//   give it a mode so it's either reading or appending - must read to end to start appending again
type WAL struct {
	BaseService

	group *auto.Group
	light bool // ignore block parts

	logger *zap.Logger
}

func NewWAL(logger *zap.Logger, walDir string, light bool) (*WAL, error) {
	group, err := auto.OpenGroup(walDir + "/wal")
	if err != nil {
		return nil, err
	}
	wal := &WAL{
		group:  group,
		light:  light,
		logger: logger,
	}
	wal.BaseService = *NewBaseService(logger, "WAL", wal)
	_, err = wal.Start()
	return wal, err
}

func (wal *WAL) OnStart() error {
	wal.BaseService.OnStart()
	size, err := wal.group.Head.Size()
	if err != nil {
		return err
	} else if size == 0 {
		wal.writeHeight(1)
	}
	_, err = wal.group.Start()
	return err
}

func (wal *WAL) OnStop() {
	wal.BaseService.OnStop()
	wal.group.Stop()
}

// called in newStep and for each pass in receiveRoutine
func (wal *WAL) Save(wmsg WALMessage) {
	if wal == nil {
		return
	}
	if wal.light {
		// in light mode we only write new steps, timeouts, and our own votes (no proposals, block parts)
		if mi, ok := wmsg.(*msgInfo); ok {
			if mi.PeerKey != "" {
				return
			}
		}
	}
	// Write #HEIGHT: XYZ if new height
	if edrs, ok := wmsg.(*agtypes.EventDataRoundState); ok {
		if edrs.Step == csspb.RoundStepType_NewHeight.CString() {
			wal.writeHeight(edrs.Height)
		}
	}
	// Write the wal message
	twm := GenTimedWALMessage(wmsg)
	wmsgBytes, err := json.Marshal(&twm)
	if err != nil {
		PanicQ(Fmt("Error writing msg to msgpack consensus wal. Error: %v \n\nMessage: %v", err, wmsg))
	}
	if err := wal.group.WriteLine(string(wmsgBytes)); err != nil {
		PanicQ(Fmt("Error writing msg to consensus wal. Error: %v \n\nMessage: %v", err, wmsg))
	}
	// TODO: only flush when necessary
	if err := wal.group.Flush(); err != nil {
		PanicQ(Fmt("Error flushing consensus wal buf to file. Error: %v \n", err))
	}
}

func (wal *WAL) writeHeight(height def.INT) {
	wal.group.WriteLine(Fmt("#HEIGHT: %v", height))

	// TODO: only flush when necessary
	if err := wal.group.Flush(); err != nil {
		PanicQ(Fmt("Error flushing consensus wal buf to file. Error: %v \n", err))
	}
}
