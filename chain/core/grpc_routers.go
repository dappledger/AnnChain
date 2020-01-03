package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	grpc2 "github.com/dappledger/AnnChain/chain/proto"
	"github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/eth/common/hexutil"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"

	"github.com/dappledger/AnnChain/gemmill/protos/rpc"
	pbtypes "github.com/dappledger/AnnChain/gemmill/protos/types"
)

func (g *grpcHandler) Status(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultStatus, error) {
	var (
		latestBlockHash []byte
		latestAppHash   []byte
		latestBlockTime int64
	)
	latestHeight := g.node.Angine.Height()
	if latestHeight != 0 {
		latestBlockMeta, err := g.node.Angine.GetBlockMeta(latestHeight)
		if err != nil {
			return nil, err
		}
		latestBlockHash = latestBlockMeta.Hash
		latestAppHash = latestBlockMeta.Header.AppHash
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	status := &rpc.ResultStatus{
		NodeInfo:          g.node.Angine.GetNodeInfo().ToPbData(),
		LatestBlockHash:   latestBlockHash,
		LatestAppHash:     latestAppHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
	}
	if g.node.Angine.PrivValidator() != nil {
		status.PubKey = g.node.Angine.PrivValidator().PubKey.ToPbData()
	}

	return status, nil
}
func (g *grpcHandler) Genesis(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultGenesis, error) {
	return &rpc.ResultGenesis{Genesis: g.node.GenesisDoc.ToPbData()}, nil
}
func (g *grpcHandler) Health(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultHealthInfo, error) {
	return &rpc.ResultHealthInfo{Status: int32(g.node.HealthStatus())}, nil
}
func (g *grpcHandler) Block(ctx context.Context, req *grpc2.CmdBlock) (*rpc.ResultBlock, error) {
	if req == nil {
		return nil, errors.New("miss request")
	}
	height := req.Height
	if height == 0 {
		return nil, fmt.Errorf("height must be greater than 0")
	}
	if height > g.node.Angine.Height() {
		return nil, fmt.Errorf("height must be less than the current blockchain height")
	}
	res := rpc.ResultBlock{}
	block, blockMeata, err := g.node.Angine.GetBlock(height)
	if block != nil {
		res.Block = block.ToPbData()
	}
	if blockMeata != nil {
		res.BlockMeta = blockMeata.ToPbData()
	}

	return &res, err
}
func (g *grpcHandler) BlockchainInfo(ctx context.Context, req *grpc2.CmdBlockchainInfo) (*rpc.ResultBlockchainInfo, error) {
	if req == nil {
		return nil, errors.New("miss request")
	}
	minHeight, maxHeight := req.MinHeight, req.MaxHeight

	if minHeight > maxHeight {
		return nil, fmt.Errorf("maxHeight has to be bigger than minHeight")
	}

	blockStoreHeight := g.node.Angine.Height()
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
		blockMeta, err := g.node.Angine.GetBlockMeta(height)
		if err != nil {
			return nil, err
		}
		blockMetas = append(blockMetas, blockMeta.ToPbData())
	}
	return &rpc.ResultBlockchainInfo{LastHeight: blockStoreHeight, BlockMetas: blockMetas}, nil
}
func (g *grpcHandler) DumpConsensusState(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultDumpConsensusState, error) {
	res := rpc.ResultDumpConsensusState{}
	res.RoundState, res.PeerRoundStates = g.node.Angine.GetConsensusStateInfo()
	return &res, nil
}
func (g *grpcHandler) UnconfirmedTxs(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultUnconfirmedTxs, error) {

	res := rpc.ResultUnconfirmedTxs{}
	txs := g.node.Angine.GetUnconfirmedTxs()
	for _, v := range txs {
		res.Txs = append(res.Txs, v)
	}
	res.N = int64(len(res.Txs))
	return &res, nil
}

func (g *grpcHandler) NumUnconfirmedTxs(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultUnconfirmedTxs, error) {
	return &rpc.ResultUnconfirmedTxs{N: int64(g.node.Angine.GetNumUnconfirmedTxs()), Txs: nil}, nil
}

func (g *grpcHandler) NumArchivedBlocks(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultNumArchivedBlocks, error) {

	return &rpc.ResultNumArchivedBlocks{Num: g.node.Angine.OriginHeight()}, nil
}

func (g *grpcHandler) UnsafeFlushMempool(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultUnsafeFlushMempool, error) {
	g.node.Angine.FlushMempool()
	return &rpc.ResultUnsafeFlushMempool{}, nil
}

func (g *grpcHandler) BroadcastTx(ctx context.Context, req *grpc2.CmdBroadcastTx) (*rpc.ResultBroadcastTx, error) {
	if req == nil || req.Tx == nil {
		return nil, errors.New("miss request")
	}
	if err := g.node.Application.CheckTx(req.Tx); err != nil {
		return nil, err
	}
	if err := g.node.Angine.BroadcastTx(req.Tx); err != nil {
		return nil, err
	}

	hash := gtypes.Tx(req.Tx).Hash()
	return &rpc.ResultBroadcastTx{TxHash: hexutil.Encode(hash), Code: 0}, nil
}

func (g *grpcHandler) BroadcastTxCommit(ctx context.Context, req *grpc2.CmdBroadcastTx) (*rpc.ResultBroadcastTxCommit, error) {
	if req == nil || req.Tx == nil {
		return nil, errors.New("miss request")
	}
	if err := g.node.Application.CheckTx(req.Tx); err != nil {
		return nil, err
	}
	if err := g.node.Angine.BroadcastTxCommit(req.Tx); err != nil {
		return nil, err
	}

	hash := gtypes.Tx(req.Tx).Hash()
	return &rpc.ResultBroadcastTxCommit{TxHash: hexutil.Encode(hash), Code: 0}, nil
}

func (g *grpcHandler) QueryTx(ctx context.Context, req *grpc2.CmdQuery) (*rpc.ResultNumLimitTx, error) {
	if req == nil || req.Query == nil {
		return nil, errors.New("miss request")
	}
	query := req.Query
	kind := query[0]
	switch kind {
	case types.QueryTxLimit:
		balance, err := g.node.Angine.Query(kind, query)
		num, _ := balance.(uint64)
		return &rpc.ResultNumLimitTx{Num: num}, err
	default:
		return nil, errors.New("unexpected query no")
	}
}

func (g *grpcHandler) Query(ctx context.Context, req *grpc2.CmdQuery) (*rpc.ResultQuery, error) {
	if req == nil || req.Query == nil {
		return nil, errors.New("miss request")
	}
	result := g.node.Application.Query(req.Query)
	return &rpc.ResultQuery{Result: result.ToPbData()}, nil
}

func (g *grpcHandler) GetTransactionByHash(ctx context.Context, req *grpc2.CmdHash) (*rpc.ResultQuery, error) {
	if req == nil {
		return nil, errors.New("miss request")
	}
	query := append([]byte{types.QueryType_TxRaw, gtypes.QueryTx}, req.Hash...)
	result := g.node.Application.Query(query)
	return &rpc.ResultQuery{Result: result.ToPbData()}, nil
}

func (g *grpcHandler) Info(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultInfo, error) {
	info := g.node.Application.Info()
	res := info.ToPbData()
	return res, nil
}

func (g *grpcHandler) Validators(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultValidators, error) {
	_, vs := g.node.Angine.GetValidators()
	validators := gtypes.MakeResultValidators(vs.Validators)
	result := &rpc.ResultValidators{
		BlockHeight: g.node.Angine.Height(),
	}
	for _, v := range validators {
		result.Validators = append(result.Validators, v.ToPbData())
	}
	return result, nil
}

func (g *grpcHandler) CoreVersion(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultCoreVersion, error) {
	appInfo := g.node.Application.Info()
	vs := strings.Split(types.GetCommitVersion(), "-")
	res := rpc.ResultCoreVersion{
		Version:    vs[0],
		AppName:    types.AppName(),
		AppVersion: appInfo.Version,
	}
	if len(vs) > 1 {
		res.Hash = vs[1]
	}

	return &res, nil
}
func (g *grpcHandler) LastHeight(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultLastHeight, error) {
	blockStoreHeight := g.node.Angine.Height()
	res := rpc.ResultLastHeight{
		LastHeight: blockStoreHeight,
	}
	return &res, nil
}
func (g *grpcHandler) ZaSurveillance(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultSurveillance, error) {
	bcHeight := g.node.Angine.Height()

	var totalNumTxs, txAvg int64
	if bcHeight >= 2 {
		startHeight := bcHeight - 200
		if startHeight < 1 {
			startHeight = 1
		}
		eBlock, _, err := g.node.Angine.GetBlock(bcHeight)
		if err != nil {
			return nil, err
		}
		endTime := eBlock.Header.Time
		sBlock, _, err := g.node.Angine.GetBlock(startHeight)
		if err != nil {
			return nil, err
		}
		startTime := sBlock.Header.Time
		totalNumTxs += int64(sBlock.Header.NumTxs)
		dura := endTime.Sub(startTime)
		for height := startHeight + 1; height < bcHeight; height++ {
			block, _, err := g.node.Angine.GetBlock(height)
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
	for _, oth := range g.node.NodeInfo().Other {
		if strings.HasPrefix(oth, "node_start_at") {
			ts, err := strconv.ParseInt(string(oth[14:]), 10, 64)
			if err != nil {
				return nil, err
			}
			runningTime = time.Duration(time.Now().Unix() - ts)
		}
	}

	_, vals := g.node.Angine.GetValidators()

	res := rpc.ResultSurveillance{
		Height:        bcHeight,
		NanoSecsPerTx: txAvg,
		Addr:          g.node.NodeInfo().RemoteAddr,
		IsValidator:   g.node.Angine.IsNodeValidator(g.node.NodeInfo().PubKey),
		NumValidators: int64(len(vals.Validators)),
		NumPeers:      int64(g.node.Angine.GetNumPeers()),
		RunningTime:   runningTime.Nanoseconds(),
		PubKey:        g.node.NodeInfo().PubKey.KeyString(),
	}
	return &res, nil
}

func (g *grpcHandler) NetInfo(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultNetInfo, error) {
	res := rpc.ResultNetInfo{}
	var peers []*gtypes.Peer
	res.Listening, res.Listeners, peers = g.node.Angine.GetP2PNetInfo()
	for _, v := range peers {
		res.Peers = append(res.Peers, v.ToPbData())
	}
	return &res, nil
}

func (g *grpcHandler) Blacklist(ctx context.Context, req *grpc2.EmptyRequest) (*rpc.ResultRefuseList, error) {
	return &rpc.ResultRefuseList{Result: g.node.Angine.GetBlacklist()}, nil
}
