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
	"time"

	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-crypto"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-p2p"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-rpc/types"
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-wire"
)

type ResultBlockchainInfo struct {
	LastHeight int          `json:"last_height"`
	BlockMetas []*BlockMeta `json:"block_metas"`
}

type ResultGenesis struct {
	Genesis *GenesisDoc `json:"genesis"`
}

type ResultBlock struct {
	BlockMeta *BlockMeta `json:"block_meta"`
	Block     *Block     `json:"block"`
}

type ResultShards struct {
	Names []string `json:"names"`
}

type ResultStatus struct {
	NodeInfo          *p2p.NodeInfo `json:"node_info"`
	PubKey            crypto.PubKey `json:"pub_key"`
	LatestBlockHash   []byte        `json:"latest_block_hash"`
	LatestAppHash     []byte        `json:"latest_app_hash"`
	LatestBlockHeight int           `json:"latest_block_height"`
	LatestBlockTime   int64         `json:"latest_block_time"` // nano
}

type ResultNetInfo struct {
	Listening bool     `json:"listening"`
	Listeners []string `json:"listeners"`
	Peers     []*Peer  `json:"peers"`
}

type ResultDialSeeds struct {
}

type Peer struct {
	p2p.NodeInfo     `json:"node_info"`
	IsOutbound       bool                 `json:"is_outbound"`
	ConnectionStatus p2p.ConnectionStatus `json:"connection_status"`
}

type ResultValidators struct {
	BlockHeight int          `json:"block_height"`
	Validators  []*Validator `json:"validators"`
}

type ResultDumpConsensusState struct {
	RoundState      string   `json:"round_state"`
	PeerRoundStates []string `json:"peer_round_states"`
}

type ResultBroadcastTx struct {
	Code CodeType `json:"code"`
	Data []byte   `json:"data"`
	Log  string   `json:"log"`
}

type ResultRequestSpecialOP struct {
	Code CodeType `json:"code"`
	Data []byte   `json:"data"`
	Log  string   `json:"log"`
}

type ResultBroadcastTxCommit struct {
	Code CodeType `json:"code"`
	Data []byte   `json:"data"`
	Log  string   `json:"log"`
}

type ResultUnconfirmedTxs struct {
	N   int  `json:"n_txs"`
	Txs []Tx `json:"txs"`
}

type ResultInfo struct {
	Data             string `json:"data"`
	Version          string `json:"version"`
	LastBlockHeight  uint64 `json:"last_block_height"`
	LastBlockAppHash []byte `json:"last_block_app_hash"`
}

type ResultQuery struct {
	Result Result `json:"result"`
}

type ResultRefuseList struct {
	Result []string `json:"result"`
}

type ResultUnsafeFlushMempool struct{}

type ResultUnsafeSetConfig struct{}

type ResultUnsafeProfile struct{}

type ResultSubscribe struct {
}

type ResultUnsubscribe struct {
}

type ResultEvent struct {
	Name string      `json:"name"`
	Data TMEventData `json:"data"`
}

type ResultSurveillance struct {
	NanoSecsPerTx time.Duration
	Height        int
	Addr          string
	IsValidator   bool
	NumValidators int
	NumPeers      int
	RunningTime   time.Duration
	PubKey        string
}

type ResultCoreVersion struct {
	Version    string `json:"version"`
	AppName    string `json:"appname"`
	AppVersion string `json:"appversion"`
	Hash       string `json:"hash"`
}

//----------------------------------------
// response & result types

const (
	// 0x0 bytes are for the blockchain
	ResultTypeGenesis        = byte(0x01)
	ResultTypeBlockchainInfo = byte(0x02)
	ResultTypeBlock          = byte(0x03)

	// 0x2 bytes are for the network
	ResultTypeStatus    = byte(0x20)
	ResultTypeNetInfo   = byte(0x21)
	ResultTypeDialSeeds = byte(0x22)
	ResultTypeShards    = byte(0x23)

	// 0x1  bytes are for refuseList
	ResultTypeRefuseList = byte(0x10)

	// 0x4 bytes are for the consensus
	ResultTypeValidators         = byte(0x40)
	ResultTypeDumpConsensusState = byte(0x41)

	// 0x6 bytes are for txs / the application
	ResultTypeBroadcastTx       = byte(0x60)
	ResultTypeUnconfirmedTxs    = byte(0x61)
	ResultTypeBroadcastTxCommit = byte(0x62)
	ResultTypeRequestSpecialOP  = byte(0x63)

	// 0x7 bytes are for querying the application
	ResultTypeQuery = byte(0x70)
	ResultTypeInfo  = byte(0x71)

	// 0x8 bytes are for events
	ResultTypeSubscribe   = byte(0x80)
	ResultTypeUnsubscribe = byte(0x81)
	ResultTypeEvent       = byte(0x82)

	// 0xa bytes for testing
	ResultTypeUnsafeSetConfig        = byte(0xa0)
	ResultTypeUnsafeStartCPUProfiler = byte(0xa1)
	ResultTypeUnsafeStopCPUProfiler  = byte(0xa2)
	ResultTypeUnsafeWriteHeapProfile = byte(0xa3)
	ResultTypeUnsafeFlushMempool     = byte(0xa4)
	ResultTypeCoreVersion            = byte(0xaf)

	// 0x9 bytes are for za_surveillance
	ResultTypeSurveillance = byte(0x90)
)

type RPCResult interface {
	rpctypes.Result
}

// for wire.readReflect
var _ = wire.RegisterInterface(
	struct{ RPCResult }{},
	wire.ConcreteType{&ResultGenesis{}, ResultTypeGenesis},
	wire.ConcreteType{&ResultBlockchainInfo{}, ResultTypeBlockchainInfo},
	wire.ConcreteType{&ResultBlock{}, ResultTypeBlock},
	wire.ConcreteType{&ResultStatus{}, ResultTypeStatus},
	wire.ConcreteType{&ResultShards{}, ResultTypeShards},
	wire.ConcreteType{&ResultNetInfo{}, ResultTypeNetInfo},
	wire.ConcreteType{&ResultDialSeeds{}, ResultTypeDialSeeds},
	wire.ConcreteType{&ResultValidators{}, ResultTypeValidators},
	wire.ConcreteType{&ResultDumpConsensusState{}, ResultTypeDumpConsensusState},
	wire.ConcreteType{&ResultBroadcastTx{}, ResultTypeBroadcastTx},
	wire.ConcreteType{&ResultBroadcastTxCommit{}, ResultTypeBroadcastTxCommit},
	wire.ConcreteType{&ResultRequestSpecialOP{}, ResultTypeRequestSpecialOP},
	wire.ConcreteType{&ResultUnconfirmedTxs{}, ResultTypeUnconfirmedTxs},
	wire.ConcreteType{&ResultSubscribe{}, ResultTypeSubscribe},
	wire.ConcreteType{&ResultUnsubscribe{}, ResultTypeUnsubscribe},
	wire.ConcreteType{&ResultEvent{}, ResultTypeEvent},
	wire.ConcreteType{&ResultUnsafeSetConfig{}, ResultTypeUnsafeSetConfig},
	wire.ConcreteType{&ResultUnsafeProfile{}, ResultTypeUnsafeStartCPUProfiler},
	wire.ConcreteType{&ResultUnsafeProfile{}, ResultTypeUnsafeStopCPUProfiler},
	wire.ConcreteType{&ResultUnsafeProfile{}, ResultTypeUnsafeWriteHeapProfile},
	wire.ConcreteType{&ResultUnsafeFlushMempool{}, ResultTypeUnsafeFlushMempool},
	wire.ConcreteType{&ResultQuery{}, ResultTypeQuery},
	wire.ConcreteType{&ResultInfo{}, ResultTypeInfo},
	wire.ConcreteType{&ResultSurveillance{}, ResultTypeSurveillance},
	wire.ConcreteType{&ResultRefuseList{}, ResultTypeRefuseList},
	wire.ConcreteType{&ResultCoreVersion{}, ResultTypeCoreVersion},
)
