package types

import (
	"bytes"
	"fmt"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-merkle"
)

func (p *Part) MerkleProof() merkle.SimpleProof {
	return merkle.SimpleProof{
		Aunts: p.GetProof().GetBytes(),
	}
}

///////////////////////////////////////////////////////////////////////////////////

func (psh *PartSetHeader) CString() string {
	if psh == nil {
		return ""
	}
	return fmt.Sprintf("%v:%X", psh.Total, Fingerprint(psh.Hash))
}

func (psh *PartSetHeader) IsZero() bool {
	if psh == nil {
		return true
	}
	return psh.Total == 0
}

func (psh *PartSetHeader) Equals(other *PartSetHeader) bool {
	if psh == other {
		return true
	}
	if psh == nil || other == nil {
		return false
	}
	return psh.Total == other.Total && bytes.Equal(psh.Hash, other.Hash)
}
