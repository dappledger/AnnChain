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

package node

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dappledger/AnnChain/ann-module/lib/go-crypto"

	at "github.com/dappledger/AnnChain/angine/types"
	rpc "github.com/dappledger/AnnChain/ann-module/lib/go-rpc/server"
	"github.com/dappledger/AnnChain/genesis/chain/version"
)

const ChainIDArg = "chainid"

// RPCNode define the node's abilities provided for rpc calls
type RPCNode interface {
	Height() int
	GetBlock(height int) (*at.Block, *at.BlockMeta)
	BroadcastTx(tx []byte) error
	BroadcastTxCommit(tx []byte) error
	FlushMempool()
	GetValidators() (int, []*at.Validator)
	GetP2PNetInfo() (bool, []string, []*at.Peer)
	GetNumPeers() int
	GetConsensusStateInfo() (string, []string)
	GetNumUnconfirmedTxs() int
	GetUnconfirmedTxs() []at.Tx
	IsNodeValidator(pub crypto.PubKey) bool
	GetBlacklist() []string
}

type rpcHandler struct {
	node *Node
}

var (
	ErrInvalidChainID = fmt.Errorf("no such chain id")
)

func newRPCHandler(n *Node) *rpcHandler {
	return &rpcHandler{node: n}
}

func (n *Node) rpcRoutes() map[string]*rpc.RPCFunc {
	h := newRPCHandler(n)
	return map[string]*rpc.RPCFunc{

		// info API
		"status":               rpc.NewRPCFunc(h.Status, argsWithChainID("")),
		"net_info":             rpc.NewRPCFunc(h.NetInfo, argsWithChainID("")),
		"blockchain":           rpc.NewRPCFunc(h.BlockchainInfo, argsWithChainID("minHeight,maxHeight")),
		"genesis":              rpc.NewRPCFunc(h.Genesis, argsWithChainID("")),
		"block":                rpc.NewRPCFunc(h.Block, argsWithChainID("height")),
		"validators":           rpc.NewRPCFunc(h.Validators, argsWithChainID("")),
		"dump_consensus_state": rpc.NewRPCFunc(h.DumpConsensusState, argsWithChainID("")),
		"unconfirmed_txs":      rpc.NewRPCFunc(h.UnconfirmedTxs, argsWithChainID("")),
		"za_surveillance":      rpc.NewRPCFunc(h.ZaSurveillance, argsWithChainID("")),
		"core_version":         rpc.NewRPCFunc(h.CoreVersion, argsWithChainID("")),
		"info":                 rpc.NewRPCFunc(h.Info, argsWithChainID("")),

		// Query RPC
		"query_nonce":                       rpc.NewRPCFunc(h.QueryNonce, argsWithChainID("address")),
		"query_account":                     rpc.NewRPCFunc(h.QueryAccount, argsWithChainID("address")),
		"query_ledgers":                     rpc.NewRPCFunc(h.QueryLedgers, argsWithChainID("order,limit,cursor")),
		"query_ledger":                      rpc.NewRPCFunc(h.QueryLedger, argsWithChainID("height")),
		"query_payments":                    rpc.NewRPCFunc(h.QueryPayments, argsWithChainID("order,limit,cursor")),
		"query_account_payments":            rpc.NewRPCFunc(h.QueryAccountPayments, argsWithChainID("address,order,limit,cursor")),
		"query_payment":                     rpc.NewRPCFunc(h.QueryPayment, argsWithChainID("txhash")),
		"query_transactions":                rpc.NewRPCFunc(h.QueryTransactions, argsWithChainID("order,limit,cursor")),
		"query_transaction":                 rpc.NewRPCFunc(h.QueryTransaction, argsWithChainID("txhash")),
		"query_account_transactions":        rpc.NewRPCFunc(h.QueryAccountTransactions, argsWithChainID("address,order,limit,cursor")),
		"query_contract":                    rpc.NewRPCFunc(h.QueryDoContract, argsWithChainID("byte[]")),
		"query_contract_exist":              rpc.NewRPCFunc(h.QueryContractExist, argsWithChainID("address")),
		"query_receipt":                     rpc.NewRPCFunc(h.QueryReceipt, argsWithChainID("txhash")),
		"query_account_managedatas":         rpc.NewRPCFunc(h.QueryAccountManagedatas, argsWithChainID("address,order,limit,cursor")),
		"query_account_managedata":          rpc.NewRPCFunc(h.QueryAccountManagedata, argsWithChainID("address,key")),
		"query_account_category_managedata": rpc.NewRPCFunc(h.QueryAccountCategoryManagedata, argsWithChainID("address,category")),
		"query_ledger_transactions":         rpc.NewRPCFunc(h.QueryLedgerTransactions, argsWithChainID("height,order,limit,cursor")),

		//Execute RPC
		"create_account":   rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
		"payment":          rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
		"manage_data":      rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
		"create_contract":  rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
		"execute_contract": rpc.NewRPCFunc(h.BroadcastTxCommit, argsWithChainID("tx")),
	}
}

func (h *rpcHandler) Status() (interface{}, at.CodeType, error) {
	var (
		latestBlockMeta *at.BlockMeta
		latestBlockHash []byte
		latestAppHash   []byte
		latestBlockTime int64
	)
	latestHeight := h.node.Angine.Height()
	if latestHeight != 0 {
		_, latestBlockMeta = h.node.Angine.GetBlock(latestHeight)
		latestBlockHash = latestBlockMeta.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}

	status := &at.ResultStatus{
		NodeInfo:          h.node.Angine.GetNodeInfo(),
		PubKey:            h.node.Angine.PrivValidator().PubKey,
		LatestBlockHash:   latestBlockHash,
		LatestAppHash:     latestAppHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}

	return &status, at.CodeType_OK, nil
}

func (h *rpcHandler) NetInfo() (interface{}, at.CodeType, error) {
	res := at.ResultNetInfo{}
	res.Listening, res.Listeners, res.Peers = h.node.Angine.GetP2PNetInfo()
	return &res, at.CodeType_OK, nil

}
func (h *rpcHandler) BlockchainInfo(minHeight, maxHeight int) (interface{}, at.CodeType, error) {
	if minHeight > maxHeight {
		return "", at.CodeType_BaseInvalidInput, errors.New("maxHeight has to be bigger than minHeight")
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
	blockMetas := []*at.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		_, blockMeta := h.node.Angine.GetBlock(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	blockChainInfo := &at.ResultBlockchainInfo{LastHeight: blockStoreHeight, BlockMetas: blockMetas}
	return &blockChainInfo, at.CodeType_OK, nil
}

func (h *rpcHandler) Genesis() (interface{}, at.CodeType, error) {
	gesesis := &at.ResultGenesis{Genesis: h.node.GenesisDoc}

	return &gesesis, at.CodeType_OK, nil
}

func (h *rpcHandler) Block(height int) (interface{}, at.CodeType, error) {
	if height == 0 {
		return "", at.CodeType_BaseInvalidInput, errors.New("height must be greater than 0")
	}
	if height > h.node.Angine.Height() {
		return "", at.CodeType_BaseInvalidInput, errors.New("height must be less than the current blockchain height")
	}
	res := at.ResultBlock{}
	res.Block, res.BlockMeta = h.node.Angine.GetBlock(height)

	return &res, at.CodeType_OK, nil
}

func (h *rpcHandler) Validators() (interface{}, at.CodeType, error) {
	height, vs := h.node.Angine.GetValidators()
	validators := &at.ResultValidators{
		Validators:  vs,
		BlockHeight: height,
	}
	return &validators, at.CodeType_OK, nil
}

func (h *rpcHandler) DumpConsensusState() (interface{}, at.CodeType, error) {
	dumpConsensusState := at.ResultDumpConsensusState{}
	dumpConsensusState.RoundState, dumpConsensusState.PeerRoundStates = h.node.Angine.GetConsensusStateInfo()

	return &dumpConsensusState, at.CodeType_OK, nil
}

func (h *rpcHandler) UnconfirmedTxs() (interface{}, at.CodeType, error) {
	unConfirmedTxs := at.ResultUnconfirmedTxs{}
	unConfirmedTxs.Txs = h.node.Angine.GetUnconfirmedTxs()
	unConfirmedTxs.N = len(unConfirmedTxs.Txs)

	return &unConfirmedTxs, at.CodeType_OK, nil
}

func (h *rpcHandler) ZaSurveillance() (interface{}, at.CodeType, error) {
	bcHeight := h.node.Angine.Height()

	var totalNumTxs, txAvg int64
	if bcHeight >= 2 {
		startHeight := bcHeight - 200
		if startHeight < 1 {
			startHeight = 1
		}
		eBlock, _ := h.node.Angine.GetBlock(bcHeight)
		endTime := eBlock.Header.Time
		sBlock, _ := h.node.Angine.GetBlock(startHeight)
		startTime := sBlock.Header.Time
		totalNumTxs += int64(sBlock.Header.NumTxs)
		dura := endTime.Sub(startTime)
		for height := startHeight + 1; height < bcHeight; height++ {
			block, _ := h.node.Angine.GetBlock(height)
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
				return "", at.CodeType_BaseInvalidInput, err
			}
			runningTime = time.Duration(time.Now().Unix() - ts)
		}
	}

	_, vals := h.node.Angine.GetValidators()

	res := at.ResultSurveillance{
		Height:        bcHeight,
		NanoSecsPerTx: time.Duration(txAvg),
		Addr:          h.node.NodeInfo().RemoteAddr,
		IsValidator:   h.node.Angine.IsNodeValidator(h.node.NodeInfo().PubKey),
		NumValidators: len(vals),
		NumPeers:      h.node.Angine.GetNumPeers(),
		RunningTime:   runningTime,
		PubKey:        h.node.NodeInfo().PubKey.KeyString(),
	}

	return &res, at.CodeType_OK, nil
}

func (h *rpcHandler) CoreVersion() (interface{}, at.CodeType, error) {
	appInfo := h.node.Application.Info()
	vs := strings.Split(version.GetCommitVersion(), "-")
	res := at.ResultCoreVersion{
		Version:    vs[0],
		AppName:    version.AppName(),
		AppVersion: appInfo.Version,
	}
	if len(vs) > 1 {
		res.Hash = vs[1]
	}

	return &res, at.CodeType_OK, nil
}

func (h *rpcHandler) Info() (interface{}, at.CodeType, error) {
	res := h.node.Application.Info()

	return &res, at.CodeType_OK, nil
}

func (h *rpcHandler) BroadcastTxCommit(tx []byte) ([]byte, at.CodeType, error) {

	if ret := h.node.Application.CheckTx(tx); ret.IsErr() {
		return nil, ret.Code, errors.New(ret.Log)
	}
	result, err := h.node.Angine.BroadcastTxCommit(tx)

	if err != nil {
		return nil, result.Code, err
	}

	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}

	return nil, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryNonce(address string) (uint64, at.CodeType, error) {
	result := h.node.Application.QueryNonce(address)
	if result.Code != at.CodeType_OK {
		return 0, result.Code, errors.New(result.Log)
	}
	return binary.BigEndian.Uint64(result.Data), at.CodeType_OK, nil
}

func (h *rpcHandler) QueryAccount(address string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryAccount(address)

	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryLedgers(order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryLedgers(order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}

	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryLedger(height uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryLedger(height)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}

	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryPayments(order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryPayments(order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}

	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryAccountPayments(address string, order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryAccountPayments(address, order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryPayment(txhash string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryPayment(txhash)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryTransactions(order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryTransactions(order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryTransaction(txhash string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryTransaction(txhash)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryAccountTransactions(address string, order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryAccountTransactions(address, order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryLedgerTransactions(height uint64, order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryLedgerTransactions(height, order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryDoContract(query []byte) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryDoContract(query)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryContractExist(address string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryContractExist(address)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryReceipt(txhash string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryReceipt(txhash)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryAccountManagedatas(address string, order string, limit uint64, cursor uint64) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryAccountManagedatas(address, order, limit, cursor)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryAccountManagedata(address string, key string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryAccountManagedata(address, key)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func (h *rpcHandler) QueryAccountCategoryManagedata(address string, category string) (interface{}, at.CodeType, error) {
	result := h.node.Application.QueryAccountCategoryManagedata(address, category)
	if result.Code != at.CodeType_OK {
		return nil, result.Code, errors.New(result.Log)
	}
	return result.Data, at.CodeType_OK, nil
}

func argsWithChainID(args string) string {
	return args
	if args == "" {
		return ChainIDArg
	}
	return ChainIDArg + "," + args
}
