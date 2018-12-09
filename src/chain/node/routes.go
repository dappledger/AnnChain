/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package node

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	rpc "github.com/dappledger/AnnChain/module/lib/go-rpc/server"
	"github.com/dappledger/AnnChain/module/xlib/def"
	//	"github.com/dappledger/AnnChain/module/lib/go-wire"
	"github.com/dappledger/AnnChain/src/chain/version"
)

const ChainIDArg = "chainid"

// RPCNode define the node's abilities provided for rpc calls
type RPCNode interface {
	GetOrg(string) (*OrgNode, error)
	Height() def.INT
	GetBlock(height def.INT) (*agtypes.BlockCache, *pbtypes.BlockMeta)
	BroadcastTx(tx []byte) error
	BroadcastTxCommit(tx []byte) error
	FlushMempool()
	GetValidators() (def.INT, []*agtypes.Validator)
	GetP2PNetInfo() (bool, []string, []*agtypes.Peer)
	GetNumPeers() int
	GetConsensusStateInfo() (string, []string)
	GetNumUnconfirmedTxs() int
	GetUnconfirmedTxs() []agtypes.Tx
	IsNodeValidator(pub crypto.PubKey) bool
	GetBlacklist() []string
}

type rpcHandler struct {
	node *Node
}

var (
	ErrInvalidChainID = errors.New("no such chain id")
	ErrMissingParams  = errors.New("missing params")
)

func newRPCHandler(n *Node) *rpcHandler {
	return &rpcHandler{node: n}
}

func (n *Node) rpcRoutes() map[string]*rpc.RPCFunc {
	h := newRPCHandler(n)
	return map[string]*rpc.RPCFunc{
		// subscribe/unsubscribe are reserved for websocket events.
		// "subscribe":   rpc.NewWSRPCFunc(SubscribeResult, argsWithChainID("event")),
		// "unsubscribe": rpc.NewWSRPCFunc(UnsubscribeResult, argsWithChainID("event")),

		// info API
		"organizations":        rpc.NewRPCFunc(h.Orgs, ""),
		"status":               rpc.NewRPCFunc(h.Status, argsWithChainID("")),
		"net_info":             rpc.NewRPCFunc(h.NetInfo, argsWithChainID("")),
		"blockchain":           rpc.NewRPCFunc(h.BlockchainInfo, argsWithChainID("minHeight,maxHeight")),
		"genesis":              rpc.NewRPCFunc(h.Genesis, argsWithChainID("")),
		"block":                rpc.NewRPCFunc(h.Block, argsWithChainID("height")),
		"validators":           rpc.NewRPCFunc(h.Validators, argsWithChainID("")),
		"dump_consensus_state": rpc.NewRPCFunc(h.DumpConsensusState, argsWithChainID("")),
		"unconfirmed_txs":      rpc.NewRPCFunc(h.UnconfirmedTxs, argsWithChainID("")),
		"num_unconfirmed_txs":  rpc.NewRPCFunc(h.NumUnconfirmedTxs, argsWithChainID("")),
		"num_archived_blocks":  rpc.NewRPCFunc(h.NumArchivedBlocks, argsWithChainID("")),
		"za_surveillance":      rpc.NewRPCFunc(h.ZaSurveillance, argsWithChainID("")),
		"core_version":         rpc.NewRPCFunc(h.CoreVersion, argsWithChainID("")),

		// broadcast API
		"broadcast_tx_commit": rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
		"broadcast_tx_sync":   rpc.NewRPCFunc(h.BroadcastTx, argsWithChainID("tx")),

		// query API
		"query":      rpc.NewRPCFunc(h.Query, argsWithChainID("query")),
		"info":       rpc.NewRPCFunc(h.Info, argsWithChainID("")),
		"event_code": rpc.NewRPCFunc(h.EventCode, argsWithChainID("code_hash")), // TODO now id is base-chain's name

		// control API
		// "dial_seeds":           rpc.NewRPCFunc(h.UnsafeDialSeeds, argsWithChainID("seeds")),
		"unsafe_flush_mempool": rpc.NewRPCFunc(h.UnsafeFlushMempool, argsWithChainID("")),
		// "unsafe_set_config":    rpc.NewRPCFunc(h.UnsafeSetConfig, argsWithChainID("type,key,value")),

		// profiler API
		// "unsafe_start_cpu_profiler": rpc.NewRPCFunc(UnsafeStartCPUProfilerResult, argsWithChainID("filename")),
		// "unsafe_stop_cpu_profiler":  rpc.NewRPCFunc(UnsafeStopCPUProfilerResult, argsWithChainID("")),
		// "unsafe_write_heap_profile": rpc.NewRPCFunc(UnsafeWriteHeapProfileResult, argsWithChainID("filename")),

		// specialOP API
		"request_special_op": rpc.NewRPCFunc(h.RequestSpecialOP, argsWithChainID("tx")),
		// "vote_special_op":    rpc.NewRPCFunc(h.VoteSpecialOP, argsWithChainID("tx")),
		"request_vote_channel": rpc.NewRPCFunc(h.RequestForVoteChannel, argsWithChainID("tx")),

		// refuse_list API
		"blacklist": rpc.NewRPCFunc(h.Blacklist, argsWithChainID("")),

		"non_empty_heights": rpc.NewRPCFunc(h.NonEmptyHeights, argsWithChainID("")),
	}
}

func (h *rpcHandler) getOrg(chainID string) (*OrgNode, error) {
	var org *OrgNode
	var err error
	if chainID == h.node.MainChainID {
		org = h.node.MainOrg
	} else {
		met := h.node.MainOrg.Application.(*Metropolis)
		org, err = met.GetOrg(chainID)
		if err != nil {
			return nil, ErrInvalidChainID
		}
	}

	return org, nil
}

func (h *rpcHandler) Orgs() (agtypes.RPCResult, error) {
	app := h.node.MainOrg.Application.(*Metropolis)
	app.Lock()
	defer app.Unlock()
	names := make([]string, 0, len(app.Orgs))
	for n := range app.Orgs {
		names = append(names, string(n))
	}
	return &agtypes.ResultOrgs{Names: names}, nil
}

func (h *rpcHandler) Status(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	var (
		latestBlockMeta *pbtypes.BlockMeta
		latestBlockHash []byte
		latestAppHash   []byte
		latestBlockTime int64
	)
	latestHeight := org.Angine.Height()
	if latestHeight != 0 {
		_, latestBlockMeta, err = org.Angine.GetBlock(latestHeight)
		if err != nil {
			return nil, err
		}
		latestBlockHash = latestBlockMeta.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTime = latestBlockMeta.Header.Time
	}

	return &agtypes.ResultStatus{
		NodeInfo:          org.Angine.GetNodeInfo(),
		PubKey:            org.Angine.PrivValidator().GetPubKey(),
		LatestBlockHash:   latestBlockHash,
		LatestAppHash:     latestAppHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (h *rpcHandler) Genesis(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return &agtypes.ResultGenesis{Genesis: org.GenesisDoc}, nil
}

func (h *rpcHandler) Block(chainID string, height def.INT) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := agtypes.ResultBlock{}
	var blockc *agtypes.BlockCache
	blockc, res.BlockMeta, err = org.Angine.GetBlock(height)
	res.Block = blockc.Block
	return &res, err
}

func (h *rpcHandler) BlockchainInfo(chainID string, minHeight, maxHeight def.INT) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	if minHeight > maxHeight {
		return nil, fmt.Errorf("maxHeight has to be bigger than minHeight")
	}

	blockStoreHeight := org.Angine.Height()
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
	blockMetas := []*pbtypes.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta, err := org.Angine.GetBlockMeta(height)
		if err != nil {
			return nil, err
		}
		blockMetas = append(blockMetas, blockMeta)
	}
	return &agtypes.ResultBlockchainInfo{LastHeight: blockStoreHeight, BlockMetas: blockMetas}, nil
}

func (h *rpcHandler) DumpConsensusState(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := agtypes.ResultDumpConsensusState{}
	res.RoundState, res.PeerRoundStates = org.Angine.GetConsensusStateInfo()
	return &res, nil
}

func (h *rpcHandler) UnconfirmedTxs(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := agtypes.ResultUnconfirmedTxs{}
	res.Txs = org.Angine.GetUnconfirmedTxs()
	res.N = len(res.Txs)
	return &res, nil
}

func (h *rpcHandler) NumUnconfirmedTxs(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return &agtypes.ResultUnconfirmedTxs{N: org.Angine.GetNumUnconfirmedTxs(), Txs: nil}, nil
}

func (h *rpcHandler) NumArchivedBlocks(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return &agtypes.ResultNumArchivedBlocks{org.Angine.OriginHeight()}, nil
}

func (h *rpcHandler) UnsafeFlushMempool(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	org.Angine.FlushMempool()
	return &agtypes.ResultUnsafeFlushMempool{}, nil
}

func (h *rpcHandler) BroadcastTx(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	if err := org.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := org.Angine.BroadcastTx(tx); err != nil {
		return nil, err
	}
	return &agtypes.ResultBroadcastTx{Code: 0}, nil
}

func (h *rpcHandler) BroadcastTxCommit(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	if err := org.Application.CheckTx(tx); err != nil {
		return nil, err
	}
	if err := org.Angine.BroadcastTxCommit(tx); err != nil {
		return nil, err
	}

	return &agtypes.ResultBroadcastTxCommit{Code: 0}, nil
}

func (h *rpcHandler) Query(chainID string, query []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return &agtypes.ResultQuery{Result: org.Application.Query(query)}, nil
}

func (h *rpcHandler) Info(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := org.Application.Info()
	return &res, nil
}

func (h *rpcHandler) EventCode(chainID string, codeHash []byte) (agtypes.RPCResult, error) {
	if len(codeHash) == 0 {
		return nil, ErrMissingParams
	}
	app := h.node.MainOrg.Application.(*Metropolis)
	ret := app.EventCodeBase.Get(codeHash)
	return &agtypes.ResultQuery{
		Result: agtypes.NewResultOK(ret, ""),
	}, nil
}

func (h *rpcHandler) Validators(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}

	_, vs := org.Angine.GetValidators()
	return &agtypes.ResultValidators{
		Validators:  vs.Validators,
		BlockHeight: org.Angine.Height(),
	}, nil
}

func (h *rpcHandler) CoreVersion(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	appInfo := org.Application.Info()
	vs := strings.Split(version.GetCommitVersion(), "-")
	res := agtypes.ResultCoreVersion{
		Version:    vs[0],
		AppName:    version.AppName(),
		AppVersion: appInfo.Version,
	}
	if len(vs) > 1 {
		res.Hash = vs[1]
	}

	return &res, nil
}

func (h *rpcHandler) RequestForVoteChannel(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return org.Angine.ProcessVoteChannel(tx)
}

func (h *rpcHandler) ZaSurveillance(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	bcHeight := org.Angine.Height()

	var totalNumTxs, txAvg int64
	if bcHeight >= 2 {
		startHeight := bcHeight - 200
		if startHeight < 1 {
			startHeight = 1
		}
		eBlock, _, err := org.Angine.GetBlock(bcHeight)
		if err != nil {
			return nil, err
		}
		endTime := agtypes.NanoToTime(eBlock.Header.Time)
		sBlock, _, err := org.Angine.GetBlock(startHeight)
		if err != nil {
			return nil, err
		}
		startTime := agtypes.NanoToTime(sBlock.Header.Time)
		totalNumTxs += int64(sBlock.Header.NumTxs)
		dura := endTime.Sub(startTime)
		for h := startHeight + 1; h < bcHeight; h++ {
			block, _, err := org.Angine.GetBlock(h)
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
				return -1, err
			}
			runningTime = time.Duration(time.Now().Unix() - ts)
		}
	}

	_, vals := org.Angine.GetValidators()

	res := agtypes.ResultSurveillance{
		Height:        bcHeight,
		NanoSecsPerTx: time.Duration(txAvg),
		Addr:          h.node.NodeInfo().RemoteAddr,
		IsValidator:   org.Angine.IsNodeValidator(&(h.node.NodeInfo().PubKey)),
		NumValidators: vals.Size(),
		NumPeers:      org.Angine.GetNumPeers(),
		RunningTime:   runningTime,
		PubKey:        h.node.NodeInfo().PubKey.KeyString(),
	}
	return &res, nil
}

func (h *rpcHandler) NetInfo(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	res := agtypes.ResultNetInfo{}
	res.Listening, res.Listeners, res.Peers = org.Angine.GetP2PNetInfo()
	return &res, nil
}

func (h *rpcHandler) Blacklist(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}
	return &agtypes.ResultRefuseList{Result: org.Angine.GetBlacklist()}, nil
}

func (h *rpcHandler) RequestSpecialOP(chainID string, tx []byte) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}

	if err := org.Angine.ProcessSpecialOP(tx); err != nil {
		res := &agtypes.ResultRequestSpecialOP{
			Code: pbtypes.CodeType_InternalError,
			Log:  err.Error(),
		}
		return res, err
	}

	return &agtypes.ResultRequestSpecialOP{
		Code: pbtypes.CodeType_OK,
	}, nil
}

// func (h *rpcHandler) VoteSpecialOP(chainID string, tx []byte) (agtypes.RPCResult, error) {
// 	org, err := h.getOrg(chainID)
// 	if err != nil {
// 		return nil, ErrInvalidChainID
// 	}
// 	if !agtypes.IsSpecialOP(tx) {
// 		return nil, fmt.Errorf("not a specialop")
// 	}

// 	cmd := new(agtypes.SpecialOPCmd)
// 	if err = wire.ReadBinaryBytes(agtypes.UnwrapTx(tx), cmd); err != nil {
// 		return nil, fmt.Errorf("error: %v", err)
// 	}
// 	res, err := org.Angine.CheckSpecialOp(cmd)
// 	if err != nil {
// 		return &agtypes.ResultRequestSpecialOP{
// 			Code: agtypes.CodeType_InternalError,
// 			Log:  err.Error(),
// 		}, err
// 	}
// 	return &agtypes.ResultRequestSpecialOP{
// 		Code: agtypes.CodeType_OK,
// 		Data: res,
// 	}, nil
// }

func (h *rpcHandler) NonEmptyHeights(chainID string) (agtypes.RPCResult, error) {
	org, err := h.getOrg(chainID)
	if err != nil {
		return nil, ErrInvalidChainID
	}

	it := org.Angine.GetNonEmptyBlockIterator()
	heights := make([]def.INT, 0)
	for it.HasMore() {
		heights = append(heights, it.Next().Header.Height)
	}

	return &agtypes.ResultNonEmptyHeights{
		Heights: heights,
	}, nil
}

// func BroadcastTxCommitResult(chainID string, tx []byte) (rpctypes.TMResult, error) {
//	shard := FindShard(chainID)
//	if err := org.Application.CheckTx(tx); err != nil {
//		return nil, err
//	}
//	org.Angine.BroadcastTxCommit(tx)
// }

func argsWithChainID(args string) string {
	if args == "" {
		return ChainIDArg
	}
	return ChainIDArg + "," + args
}
