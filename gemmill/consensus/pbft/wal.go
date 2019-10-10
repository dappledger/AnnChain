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

package pbft

import (
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-wire"
	auto "github.com/dappledger/AnnChain/gemmill/modules/go-autofile"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/types"
)

//--------------------------------------------------------
// types and functions for savings consensus messages

type TimedWALMessage struct {
	Time time.Time  `json:"time"`
	Msg  WALMessage `json:"msg"`
}

type WALMessage interface{}

var _ = wire.RegisterInterface(
	struct{ WALMessage }{},
	wire.ConcreteType{types.EventDataRoundState{}, 0x01},
	wire.ConcreteType{msgInfo{}, 0x02},
	wire.ConcreteType{timeoutInfo{}, 0x03},
)

//--------------------------------------------------------
// Simple write-ahead log

// Write ahead log writes msgs to disk before they are processed.
// Can be used for crash-recovery and deterministic replay
// TODO: currently the wal is overwritten during replay catchup
//   give it a mode so it's either reading or appending - must read to end to start appending again
type WAL struct {
	gcmn.BaseService

	group *auto.Group
	light bool // ignore block parts
}

func NewWAL(walDir string, light bool) (*WAL, error) {
	group, err := auto.OpenGroup(walDir + "/wal")
	if err != nil {
		return nil, err
	}
	wal := &WAL{
		group: group,
		light: light,
	}
	wal.BaseService = *gcmn.NewBaseService("WAL", wal)
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
		if mi, ok := wmsg.(msgInfo); ok {
			if mi.PeerKey != "" {
				return
			}
		}
	}
	// Write #HEIGHT: XYZ if new height
	if edrs, ok := wmsg.(types.EventDataRoundState); ok {
		if edrs.Step == RoundStepNewHeight.String() {
			wal.writeHeight(edrs.Height)
		}
	}
	// Write the wal message
	var wmsgBytes = wire.JSONBytes(TimedWALMessage{time.Now(), wmsg})
	err := wal.group.WriteLine(string(wmsgBytes))
	if err != nil {
		gcmn.PanicQ(gcmn.Fmt("Error writing msg to consensus wal. Error: %v \n\nMessage: %v", err, wmsg))
	}
	// TODO: only flush when necessary
	if err := wal.group.Flush(); err != nil {
		gcmn.PanicQ(gcmn.Fmt("Error flushing consensus wal buf to file. Error: %v \n", err))
	}
}

func (wal *WAL) writeHeight(height int64) {
	wal.group.WriteLine(gcmn.Fmt("#HEIGHT: %v", height))

	// TODO: only flush when necessary
	if err := wal.group.Flush(); err != nil {
		gcmn.PanicQ(gcmn.Fmt("Error flushing consensus wal buf to file. Error: %v \n", err))
	}
}
