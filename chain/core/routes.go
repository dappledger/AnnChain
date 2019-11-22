// Copyright © 2017 ZhongAn Technology
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

package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	rpc "github.com/dappledger/AnnChain/gemmill/rpc/server"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

// RPCNode define the node's abilities provided for rpc calls
type RPCNode interface {
	// GetShard(string) (*ShardNode, error)
	Height() int
	GetBlock(height int) (*gtypes.Block, *gtypes.BlockMeta, error)
	BroadcastTx(tx []byte) error
	BroadcastTxCommit(tx []byte) error
	FlushMempool()
	GetValidators() (int, []*gtypes.Validator)
	GetP2PNetInfo() (bool, []string, []*gtypes.Peer)
	GetNumPeers() int
	GetConsensusStateInfo() (string, []string)
	GetNumUnconfirmedTxs() int
	GetUnconfirmedTxs() []gtypes.Tx
	IsNodeValidator(pub crypto.PubKey) bool
	GetBlacklist() []string
}

type rpcHandler struct {
	node *Node
}

func newRPCHandler(n *Node) *rpcHandler {
	return &rpcHandler{node: n}
}

func (n *Node) rpcRoutes() map[string]*rpc.RPCFunc {
	h := newRPCHandler(n)
	return map[string]*rpc.RPCFunc{
		// subscribe/unsubscribe are reserved for websocket events.
		// "subscribe":   rpc.NewWSRPCFunc(SubscribeResult, "event"),
		// "unsubscribe": rpc.NewWSRPCFunc(UnsubscribeResult, "event"),

		// info API
		// "shards":               rpc.NewRPCFunc(h.Shards, ""),
		"status":               rpc.NewRPCFunc(h.Status, ""),
		"healthinfo":           rpc.NewRPCFunc(h.HealthInfo, ""),
		"net_info":             rpc.NewRPCFunc(h.NetInfo, ""),
		"blockchain":           rpc.NewRPCFunc(h.BlockchainInfo, "minHeight,maxHeight"),
		"genesis":              rpc.NewRPCFunc(h.Genesis, ""),
		"block":                rpc.NewRPCFunc(h.Block, "height"),
		"validators":           rpc.NewRPCFunc(h.Validators, ""),
		"dump_consensus_state": rpc.NewRPCFunc(h.DumpConsensusState, ""),
		"unconfirmed_txs":      rpc.NewRPCFunc(h.UnconfirmedTxs, ""),
		"num_unconfirmed_txs":  rpc.NewRPCFunc(h.NumUnconfirmedTxs, ""),
		"num_archived_blocks":  rpc.NewRPCFunc(h.NumArchivedBlocks, ""),
		"za_surveillance":      rpc.NewRPCFunc(h.ZaSurveillance, ""),
		"core_version":         rpc.NewRPCFunc(h.CoreVersion, ""),

		"last_height": rpc.NewRPCFunc(h.LastHeight, ""),

		// broadcast API
		"broadcast_tx_commit": rpc.NewRPCFunc(h.BroadcastTxCommit, "tx"),
		"broadcast_tx_async":  rpc.NewRPCFunc(h.BroadcastTx, "tx"),

		// query API
		"query":   rpc.NewRPCFunc(h.Query, "query"),
		"querytx": rpc.NewRPCFunc(h.QueryTx, "query"),
		"info":    rpc.NewRPCFunc(h.Info, ""),

		"transaction": rpc.NewRPCFunc(h.GetTransactionByHash, "tx"),

		// control API
		// "dial_seeds":           rpc.NewRPCFunc(h.UnsafeDialSeeds, "seeds"),
		"unsafe_flush_mempool": rpc.NewRPCFunc(h.UnsafeFlushMempool, ""),
		// "unsafe_set_config":    rpc.NewRPCFunc(h.UnsafeSetConfig, "type,key,value"),

		// profiler API
		// "unsafe_start_cpu_profiler": rpc.NewRPCFunc(UnsafeStartCPUProfilerResult, "filename"),
		// "unsafe_stop_cpu_profiler":  rpc.NewRPCFunc(UnsafeStopCPUProfilerResult, ""),
		// "unsafe_write_heap_profile": rpc.NewRPCFunc(UnsafeWriteHeapProfileResult, "filename"),

		// refuse_list API
		"blacklist": rpc.NewRPCFunc(h.Blacklist, ""),

		// sharding API
		// "shard_join": rpc.NewRPCFunc(h.ShardJoin, "gdata,cdata,sig"),
	}
}

func (h *rpcHandler) Status() (*gtypes.ResultStatus, error) {
	var (
		latestBlockHash []byte
		latestAppHash   []byte
		latestBlockTime int64
	)
	latestHeight := h.node.Angine.Height()
	if latestHeight != 0 {
		latestBlockMeta, err := h.node.Angine.GetBlockMeta(latestHeight)
		if err != nil {
			return nil, err
		}
		latestBlockHash = latestBlockMeta.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}

	return &gtypes.ResultStatus{
		NodeInfo:          h.node.Angine.GetNodeInfo(),
		PubKey:            h.node.Angine.PrivValidator().PubKey,
		LatestBlockHash:   latestBlockHash,
		LatestAppHash:     latestAppHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (h *rpcHandler) Genesis() (*gtypes.ResultGenesis, error) {
	return &gtypes.ResultGenesis{Genesis: h.node.GenesisDoc}, nil
}

func (h *rpcHandler) HealthInfo() (*gtypes.ResultHealthInfo, error) {
	return &gtypes.ResultHealthInfo{Status: h.node.HealthStatus()}, nil
}

func (h *rpcHandler) Block(height int64) (*gtypes.ResultBlock, error) {

	if height == 0 {
		return nil, fmt.Errorf("height must be greater than 0")
	}
	if height > h.node.Angine.Height() {
		return nil, fmt.Errorf("height must be less than the current blockchain height")
	}
	res := gtypes.ResultBlock{}
	var err error
	res.Block, res.BlockMeta, err = h.node.Angine.GetBlock(height)
	return &res, err
}

func (h *rpcHandler) BlockchainInfo(minHeight, maxHeight int64) (*gtypes.ResultBlockchainInfo, error) {
	if minHeight > maxHeight {
		return nil, fmt.Errorf("maxHeight has to be bigger than minHeight")
	}

	blockStoreHeight := h.node.Angine.Height()
	if maxHeight == 0 {
		maxHeight = blockStoreHeight
	} else if blockStoreHeight < maxHeight {
		maxHeight = blockStoreHeight
	}
	if minHeight == 0 {
		if maxHeight-20 > 1 {
			minHeight = maxHeight - 20
		} else {
			minHeight = 1
		}
	}
	blockMetas := []*gtypes.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta, err := h.node.Angine.GetBlockMeta(height)
		if err != nil {
			return nil, err
		}
		blockMetas = append(blockMetas, blockMeta)
	}
	return &gtypes.ResultBlockchainInfo{LastHeight: blockStoreHeight, BlockMetas: blockMetas}, nil
}

func (h *rpcHandler) DumpConsensusState() (*gtypes.ResultDumpConsensusState, error) {
	res := gtypes.ResultDumpConsensusState{}
	res.RoundState, res.PeerRoundStates = h.node.Angine.GetConsensusStateInfo()
	return &res, nil
}

func (h *rpcHandler) UnconfirmedTxs() (*gtypes.ResultUnconfirmedTxs, error) {
	res := gtypes.ResultUnconfirmedTxs{}
	res.Txs = h.node.Angine.GetUnconfirmedTxs()
	res.N = len(res.Txs)
	return &res, nil
}

func (h *rpcHandler) NumUnconfirmedTxs() (*gtypes.ResultUnconfirmedTxs, error) {
	return &gtypes.ResultUnconfirmedTxs{N: h.node.Angine.GetNumUnconfirmedTxs(), Txs: nil}, nil
}

func (h *rpcHandler) NumArchivedBlocks() (*gtypes.ResultNumArchivedBlocks, error) {
	return &gtypes.ResultNumArchivedBlocks{Num: h.node.Angine.OriginHeight()}, nil
}

func (h *rpcHandler) UnsafeFlushMempool() (*gtypes.ResultUnsafeFlushMempool, error) {
	h.node.Angine.FlushMempool()
	return &gtypes.ResultUnsafeFlushMempool{}, nil
}

func (h *rpcHandler) BroadcastTx(tx []byte) (*gtypes.ResultBroadcastTx, error) {
	if err := h.node.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := h.node.Angine.BroadcastTx(tx); err != nil {
		return nil, err
	}

	hash := gtypes.Tx(tx).Hash()
	return &gtypes.ResultBroadcastTx{TxHash: hexutil.Encode(hash), Code: 0}, nil
}

func (h *rpcHandler) BroadcastTxCommit(tx []byte) (*gtypes.ResultBroadcastTxCommit, error) {
	if err := h.node.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := h.node.Angine.BroadcastTxCommit(tx); err != nil {
		return nil, err
	}

	hash := gtypes.Tx(tx).Hash()
	return &gtypes.ResultBroadcastTxCommit{TxHash: hexutil.Encode(hash), Code: 0}, nil
}

func (h *rpcHandler) QueryTx(query []byte) (*gtypes.ResultNumLimitTx, error) {
	kind := query[0]
	switch kind {
	case types.QueryTxLimit:
		balance, err := h.node.Angine.Query(kind, query)
		num, _ := balance.(uint64)
		return &gtypes.ResultNumLimitTx{Num: num}, err
	default:
		return nil, errors.New("unexpected query no")
	}

}

func (h *rpcHandler) Query(query []byte) (*gtypes.ResultQuery, error) {
	return &gtypes.ResultQuery{Result: h.node.Application.Query(query)}, nil
}

func (h *rpcHandler) GetTransactionByHash(hash []byte) (*gtypes.ResultQuery, error) {
	query := append([]byte{types.QueryType_TxRaw, gtypes.QueryTx}, hash...)
	return &gtypes.ResultQuery{Result: h.node.Application.Query(query)}, nil
}

func (h *rpcHandler) Info() (*gtypes.ResultInfo, error) {
	res := h.node.Application.Info()
	return &res, nil
}

func (h *rpcHandler) Validators() (*gtypes.ResultValidators, error) {
	_, vs := h.node.Angine.GetValidators()
	return &gtypes.ResultValidators{
		Validators:  gtypes.MakeResultValidators(vs.Validators),
		BlockHeight: h.node.Angine.Height(),
	}, nil
}

func (h *rpcHandler) CoreVersion() (*gtypes.ResultCoreVersion, error) {
	appInfo := h.node.Application.Info()
	vs := strings.Split(types.GetCommitVersion(), "-")
	res := gtypes.ResultCoreVersion{
		Version:    vs[0],
		AppName:    types.AppName(),
		AppVersion: appInfo.Version,
	}
	if len(vs) > 1 {
		res.Hash = vs[1]
	}

	return &res, nil
}

func (h *rpcHandler) LastHeight() (*gtypes.ResultLastHeight, error) {
	blockStoreHeight := h.node.Angine.Height()
	res := gtypes.ResultLastHeight{
		LastHeight: blockStoreHeight,
	}
	return &res, nil
}

func (h *rpcHandler) ZaSurveillance() (*gtypes.ResultSurveillance, error) {
	bcHeight := h.node.Angine.Height()

	var totalNumTxs, txAvg int64
	if bcHeight >= 2 {
		startHeight := bcHeight - 200
		if startHeight < 1 {
			startHeight = 1
		}
		eBlock, _, err := h.node.Angine.GetBlock(bcHeight)
		if err != nil {
			return nil, err
		}
		endTime := eBlock.Header.Time
		sBlock, _, err := h.node.Angine.GetBlock(startHeight)
		if err != nil {
			return nil, err
		}
		startTime := sBlock.Header.Time
		totalNumTxs += int64(sBlock.Header.NumTxs)
		dura := endTime.Sub(startTime)
		for height := startHeight + 1; height < bcHeight; height++ {
			block, _, err := h.node.Angine.GetBlock(height)
			if err != nil {
				return nil, err
			}
			totalNumTxs += int64(block.Header.NumTxs)
		}
		if totalNumTxs > 0 {
			txAvg = int64(dura) / totalNumTxs
		}
	}

	var runningTime time.Duration
	for _, oth := range h.node.NodeInfo().Other {
		if strings.HasPrefix(oth, "node_start_at") {
			ts, err := strconv.ParseInt(string(oth[14:]), 10, 64)
			if err != nil {
				return nil, err
			}
			runningTime = time.Duration(time.Now().Unix() - ts)
		}
	}

	_, vals := h.node.Angine.GetValidators()

	res := gtypes.ResultSurveillance{
		Height:        bcHeight,
		NanoSecsPerTx: time.Duration(txAvg),
		Addr:          h.node.NodeInfo().RemoteAddr,
		IsValidator:   h.node.Angine.IsNodeValidator(h.node.NodeInfo().PubKey),
		NumValidators: len(vals.Validators),
		NumPeers:      h.node.Angine.GetNumPeers(),
		RunningTime:   runningTime,
		PubKey:        h.node.NodeInfo().PubKey.KeyString(),
	}
	return &res, nil
}

func (h *rpcHandler) NetInfo() (*gtypes.ResultNetInfo, error) {
	res := gtypes.ResultNetInfo{}
	res.Listening, res.Listeners, res.Peers = h.node.Angine.GetP2PNetInfo()
	return &res, nil
}

func (h *rpcHandler) Blacklist() (*gtypes.ResultRefuseList, error) {
	return &gtypes.ResultRefuseList{Result: h.node.Angine.GetBlacklist()}, nil
}
