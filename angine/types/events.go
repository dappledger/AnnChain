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

package types

import (
	"go.uber.org/zap"

	// for registering TMEventData as events.EventData
	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	"github.com/dappledger/AnnChain/ann-module/lib/go-events"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

// Functions to generate eventId strings

// Reserved
func EventStringBond() string    { return "Bond" }
func EventStringUnbond() string  { return "Unbond" }
func EventStringRebond() string  { return "Rebond" }
func EventStringDupeout() string { return "Dupeout" }
func EventStringFork() string    { return "Fork" }
func EventStringTx(tx Tx) string { return Fmt("Tx:%X", tx.Hash()) }

func EventStringNewBlock() string         { return "NewBlock" }
func EventStringNewBlockHeader() string   { return "NewBlockHeader" }
func EventStringNewRound() string         { return "NewRound" }
func EventStringNewRoundStep() string     { return "NewRoundStep" }
func EventStringTimeoutPropose() string   { return "TimeoutPropose" }
func EventStringCompleteProposal() string { return "CompleteProposal" }
func EventStringPolka() string            { return "Polka" }
func EventStringUnlock() string           { return "Unlock" }
func EventStringLock() string             { return "Lock" }
func EventStringRelock() string           { return "Relock" }
func EventStringTimeoutWait() string      { return "TimeoutWait" }
func EventStringVote() string             { return "Vote" }

func EventStringSwitchToConsensus() string { return "SwitchToConsensus" }

func EventStringHookPrevote() string   { return "Hook Prevote" }
func EventStringHookNewRound() string  { return "Hook NewRound" }
func EventStringHookPropose() string   { return "Hook Propose" }
func EventStringHookCommit() string    { return "Hook Commit" }
func EventStringHookPrecommit() string { return "Hook Precommit" }
func EventStringHookExecute() string   { return "Hook Execute" }

//----------------------------------------

// implements events.EventData
type TMEventData interface {
	events.EventData
	AssertIsTMEventData()
}

const (
	EventDataTypeNewBlock       = byte(0x01)
	EventDataTypeFork           = byte(0x02)
	EventDataTypeTx             = byte(0x03)
	EventDataTypeNewBlockHeader = byte(0x04)

	EventDataTypeSwitchToConsensus = byte(0x5)

	EventDataTypeRoundState = byte(0x11)
	EventDataTypeVote       = byte(0x12)

	EventDataTypeHookNewRound  = byte(0x21)
	EventDataTypeHookPropose   = byte(0x22)
	EventDataTypeHookPrevote   = byte(0x23)
	EventDataTypeHookPrecommit = byte(0x24)
	EventDataTypeHookCommit    = byte(0x25)
	EventDataTypeHookExecute   = byte(0x26)
)

var _ = wire.RegisterInterface(
	struct{ TMEventData }{},
	wire.ConcreteType{EventDataNewBlock{}, EventDataTypeNewBlock},
	wire.ConcreteType{EventDataNewBlockHeader{}, EventDataTypeNewBlockHeader},
	// wire.ConcreteType{EventDataFork{}, EventDataTypeFork },
	wire.ConcreteType{EventDataTx{}, EventDataTypeTx},
	wire.ConcreteType{EventDataRoundState{}, EventDataTypeRoundState},
	wire.ConcreteType{EventDataVote{}, EventDataTypeVote},

	wire.ConcreteType{EventDataSwitchToConsensus{}, EventDataTypeSwitchToConsensus},

	wire.ConcreteType{EventDataHookNewRound{}, EventDataTypeHookNewRound},
	wire.ConcreteType{EventDataHookPropose{}, EventDataTypeHookPropose},
	wire.ConcreteType{EventDataHookPrevote{}, EventDataTypeHookPrevote},
	wire.ConcreteType{EventDataHookPrecommit{}, EventDataTypeHookPrecommit},
	wire.ConcreteType{EventDataHookCommit{}, EventDataTypeHookCommit},
	wire.ConcreteType{EventDataHookExecute{}, EventDataTypeHookExecute},
)

// Most event messages are basic types (a block, a transaction)
// but some (an input to a call tx or a receive) are more exotic

type EventDataNewBlock struct {
	Block *Block `json:"block"`
}

// light weight event for benchmarking
type EventDataNewBlockHeader struct {
	Header *Header `json:"header"`
}

// All txs fire EventDataTx
type EventDataTx struct {
	Tx    Tx       `json:"tx"`
	Data  []byte   `json:"data"`
	Log   string   `json:"log"`
	Code  CodeType `json:"code"`
	Error string   `json:"error"` // this is redundant information for now
}

// NOTE: This goes into the replay WAL
type EventDataRoundState struct {
	Height int    `json:"height"`
	Round  int    `json:"round"`
	Step   string `json:"step"`

	// private, not exposed to websockets
	RoundState interface{} `json:"-"`
}

type EventDataVote struct {
	Vote *Vote
}

type EventDataSwitchToConsensus struct {
	State interface{}
}

type EventDataHookNewRound struct {
	Height int
	Round  int
	ResCh  chan NewRoundResult
}
type EventDataHookPropose struct {
	Height int
	Round  int
	Block  *Block
}
type EventDataHookPrevote struct {
	Height int
	Round  int
	Block  *Block
}
type EventDataHookPrecommit struct {
	Height int
	Round  int
	Block  *Block
}
type EventDataHookCommit struct {
	Height int
	Round  int
	Block  *Block
	ResCh  chan CommitResult
}
type EventDataHookExecute struct {
	Height int
	Round  int
	Block  *Block
	ResCh  chan ExecuteResult
}

func NewEventDataHookNewRound(height, round int) EventDataHookNewRound {
	return EventDataHookNewRound{
		Height: height,
		Round:  round,
		ResCh:  make(chan NewRoundResult, 1),
	}
}

func NewEventDataHookExecute(height, round int, block *Block) EventDataHookExecute {
	return EventDataHookExecute{
		Height: height,
		Round:  round,
		Block:  block,
		ResCh:  make(chan ExecuteResult, 1),
	}
}

func NewEventDataHookCommit(height, round int, block *Block) EventDataHookCommit {
	return EventDataHookCommit{
		Height: height,
		Round:  round,
		Block:  block,
		ResCh:  make(chan CommitResult, 1),
	}
}

func (_ EventDataNewBlock) AssertIsTMEventData()          {}
func (_ EventDataNewBlockHeader) AssertIsTMEventData()    {}
func (_ EventDataTx) AssertIsTMEventData()                {}
func (_ EventDataRoundState) AssertIsTMEventData()        {}
func (_ EventDataVote) AssertIsTMEventData()              {}
func (_ EventDataSwitchToConsensus) AssertIsTMEventData() {}

func (_ EventDataHookNewRound) AssertIsTMEventData()  {}
func (_ EventDataHookPropose) AssertIsTMEventData()   {}
func (_ EventDataHookPrevote) AssertIsTMEventData()   {}
func (_ EventDataHookPrecommit) AssertIsTMEventData() {}
func (_ EventDataHookCommit) AssertIsTMEventData()    {}
func (_ EventDataHookExecute) AssertIsTMEventData()   {}

//----------------------------------------
// Wrappers for type safety

type Fireable interface {
	events.Fireable
}

type Eventable interface {
	SetEventSwitch(EventSwitch)
}

type EventSwitch interface {
	events.EventSwitch
}

type EventCache interface {
	Fireable
	Flush()
}

func NewEventSwitch(logger *zap.Logger) EventSwitch {
	return events.NewEventSwitch(logger)
}

func NewEventCache(evsw EventSwitch) EventCache {
	return events.NewEventCache(evsw)
}

// All events should be based on this FireEvent to ensure they are TMEventData
func fireEvent(fireable events.Fireable, event string, data TMEventData) {
	if fireable != nil {
		fireable.FireEvent(event, data)
	}
}

func AddListenerForEvent(evsw EventSwitch, id, event string, cb func(data TMEventData)) {
	evsw.AddListenerForEvent(id, event, func(data events.EventData) {
		cb(data.(TMEventData))
	})

}

//--- block, tx, and vote events

func FireEventNewBlock(fireable events.Fireable, block EventDataNewBlock) {
	fireEvent(fireable, EventStringNewBlock(), block)
}

func FireEventNewBlockHeader(fireable events.Fireable, header EventDataNewBlockHeader) {
	fireEvent(fireable, EventStringNewBlockHeader(), header)
}

func FireEventVote(fireable events.Fireable, vote EventDataVote) {
	fireEvent(fireable, EventStringVote(), vote)
}

func FireEventTx(fireable events.Fireable, tx EventDataTx) {
	fireEvent(fireable, EventStringTx(tx.Tx), tx)
}

//--- EventDataRoundState events

func FireEventNewRoundStep(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringNewRoundStep(), rs)
}

func FireEventTimeoutPropose(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringTimeoutPropose(), rs)
}

func FireEventTimeoutWait(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringTimeoutWait(), rs)
}

func FireEventNewRound(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringNewRound(), rs)
}

func FireEventCompleteProposal(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringCompleteProposal(), rs)
}

func FireEventPolka(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringPolka(), rs)
}

func FireEventUnlock(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringUnlock(), rs)
}

func FireEventRelock(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringRelock(), rs)
}

func FireEventLock(fireable events.Fireable, rs EventDataRoundState) {
	fireEvent(fireable, EventStringLock(), rs)
}

func FireEventSwitchToConsensus(fireable events.Fireable) {
	fireEvent(fireable, EventStringSwitchToConsensus(), EventDataSwitchToConsensus{})
}

func FireEventHookNewRound(fireable events.Fireable, d EventDataHookNewRound) {
	fireEvent(fireable, EventStringHookNewRound(), d)
}
func FireEventHookPrevote(fireable events.Fireable, d EventDataHookPrevote) {
	fireEvent(fireable, EventStringHookPrevote(), d)
}
func FireEventHookPropose(fireable events.Fireable, d EventDataHookPropose) {
	fireEvent(fireable, EventStringHookPropose(), d)
}
func FireEventHookPrecommit(fireable events.Fireable, d EventDataHookPrecommit) {
	fireEvent(fireable, EventStringHookPrecommit(), d)
}
func FireEventHookCommit(fireable events.Fireable, d EventDataHookCommit) {
	fireEvent(fireable, EventStringHookCommit(), d)
}
func FireEventHookExecute(fireable events.Fireable, d EventDataHookExecute) {
	fireEvent(fireable, EventStringHookExecute(), d)
}
