package types

import (
	pbrpc "github.com/dappledger/AnnChain/gemmill/protos/rpc"
	pbtypes "github.com/dappledger/AnnChain/gemmill/protos/types"
)

func (v GenesisValidator) ToPbData() *pbtypes.GenesisValidator {
	pv := pbtypes.GenesisValidator{
		Pubkey:     v.PubKey.ToPbData(),
		Amount:     v.Amount,
		Name:       v.Name,
		IsCA:       v.IsCA,
		RPCAddress: v.RPCAddress,
	}
	return &pv
}

func (g *GenesisDoc) ToPbData() *pbtypes.GenesisDoc {
	if g == nil {
		return nil
	}
	pg := pbtypes.GenesisDoc{
		GenesisTime: g.GenesisTime.UnixNano() / 1e6,
		ChainID:     g.ChainID,
		AppHash:     g.AppHash,
		Plugins:     g.Plugins,
	}
	for _, v := range g.Validators {
		validator := v.ToPbData()
		pg.Validators = append(pg.Validators, validator)
	}
	return &pg
}

func (psh PartSetHeader) ToPbData() *pbtypes.PartSetHeader {
	pbPsh := pbtypes.PartSetHeader{Hash: psh.Hash, Total: int32(psh.Total)}
	return &pbPsh
}

func (id BlockID) ToPbData() *pbtypes.BlockID {
	pbId := pbtypes.BlockID{
		Hash:        id.Hash,
		PartsHeader: id.PartsHeader.ToPbData(),
	}
	return &pbId
}

func (d *Data) ToPbData() *pbtypes.Data {
	if d == nil {
		return nil
	}
	pbData := pbtypes.Data{
		Hash: d.hash,
	}
	for _, v := range d.Txs {
		pbData.Txs = append(pbData.Txs, v)
	}
	for _, v := range d.ExTxs {
		pbData.ExTxs = append(pbData.ExTxs, v)
	}
	return &pbData
}

func (v *Vote) ToPbData() *pbtypes.Vote {
	if v == nil {
		return nil
	}
	pbVote := pbtypes.Vote{
		ValidatorAddress: v.ValidatorAddress,
		ValidatorIndex:   int32(v.ValidatorIndex),
		Height:           v.Height,
		Round:            v.Round,
		Type:             pbtypes.VoteType(v.Type),
		BlockID:          v.BlockID.ToPbData(),
		Signature:        v.Signature.ToPbData(),
	}
	return &pbVote
}

func (c *Commit) ToPbData() *pbtypes.Commit {
	if c == nil {
		return nil
	}
	pbCommit := pbtypes.Commit{
		BlockID: c.BlockID.ToPbData(),
	}
	for _, v := range c.Precommits {
		pbCommit.Precommits = append(pbCommit.Precommits, v.ToPbData())
	}
	return &pbCommit
}

func (b *Block) ToPbData() *pbtypes.Block {
	if b == nil {
		return nil
	}
	pbBlock := pbtypes.Block{
		Header:     b.Header.ToPbData(),
		Data:       b.Data.ToPbData(),
		LastCommit: b.LastCommit.ToPbData(),
	}
	return &pbBlock
}

func (m *BlockMeta) ToPbData() *pbtypes.BlockMeta {
	if m == nil {
		return nil
	}
	pbBlockMeta := pbtypes.BlockMeta{
		Hash:        m.Hash,
		Header:      m.Header.ToPbData(),
		PartsHeader: m.PartsHeader.ToPbData(),
	}
	return &pbBlockMeta
}

func (r *Result) ToPbData() *pbtypes.Result {
	if r == nil {
		return nil
	}
	pbResult := pbtypes.Result{
		Code: pbtypes.CodeType(r.Code),
		Data: r.Data,
		Log:  r.Log,
	}
	return &pbResult
}

func (r *ResultInfo) ToPbData() *pbrpc.ResultInfo {
	if r == nil {
		return nil
	}
	pbResult := pbrpc.ResultInfo{
		Data:             r.Data,
		Version:          r.Version,
		LastBlockHeight:  r.LastBlockHeight,
		LastBlockAppHash: r.LastBlockAppHash,
	}
	return &pbResult
}

func (r *ResultValidator) ToPbData() *pbrpc.ResultValidator {
	if r == nil {
		return nil
	}
	pbResult := pbrpc.ResultValidator{
		Address:     r.Address,
		PubKey:      r.PubKey,
		VotingPower: r.VotingPower,
		IsCA:        r.IsCA,
		Accum:       r.Accum,
	}
	return &pbResult
}

func (p *Peer) ToPbData() *pbrpc.Peer {
	if p == nil {
		return nil
	}
	pbPeer := pbrpc.Peer{
		NodeInfo:   p.NodeInfo.ToPbData(),
		IsOutbound: p.IsOutbound,
	}
	pbPeer.ConnectionStatus = p.ConnectionStatus.ToPbData()
	return &pbPeer
}

func (h *Header) ToPbData() *pbtypes.Header {
	if h == nil {
		return nil
	}
	ph := pbtypes.Header{
		ChainID:         h.ChainID,
		Height:          h.Height,
		Time:            h.Time.UnixNano() / 1e6,
		NumTxs:          h.NumTxs,
		LastBlockID:     h.LastBlockID.ToPbData(),
		LastCommitHash:  h.LastCommitHash,
		DataHash:        h.DataHash,
		ValidatorsHash:  h.ValidatorsHash,
		AppHash:         h.AppHash,
		ReceiptsHash:    h.ReceiptsHash,
		ProposerAddress: h.ProposerAddress,
		Extra:           h.Extra,
	}
	return &ph
}