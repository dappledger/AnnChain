package types

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

func (p *Proposal) CString() string {
	if p == nil || p.GetData() == nil {
		return "nil"
	}
	pdata := p.GetData()
	return fmt.Sprintf("Proposal{%v/%v %v (%v,%v) %X}", pdata.Height, pdata.Round,
		pdata.BlockPartsHeader, pdata.POLRound, pdata.POLBlockID.CString(), p.Signature)
}

func (pdata *ProposalData) GetBytes(chainID string) (bys []byte, err error) {
	bys, err = proto.Marshal(pdata)
	if err != nil {
		return nil, err
	}
	st := SignableBase{
		ChainID: chainID,
		Data:    bys,
	}
	bys, err = proto.Marshal(&st)
	return bys, err
}
