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
	"errors"
	"fmt"
	"io"

	//. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	"github.com/dappledger/AnnChain/ann-module/lib/go-crypto"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

var (
	ErrInvalidBlockPartSignature = errors.New("Error invalid block part signature")
	ErrInvalidBlockPartHash      = errors.New("Error invalid block part hash")
)

type Proposal struct {
	Height           int              `json:"height"`
	Round            int              `json:"round"`
	BlockPartsHeader PartSetHeader    `json:"block_parts_header"`
	POLRound         int              `json:"pol_round"`    // -1 if null.
	POLBlockID       BlockID          `json:"pol_block_id"` // zero if null.
	Signature        crypto.Signature `json:"signature"`
}

// polRound: -1 if no polRound.
func NewProposal(height int, round int, blockPartsHeader PartSetHeader, polRound int, polBlockID BlockID) *Proposal {
	return &Proposal{
		Height:           height,
		Round:            round,
		BlockPartsHeader: blockPartsHeader,
		POLRound:         polRound,
		POLBlockID:       polBlockID,
	}
}

func (p *Proposal) String() string {
	return fmt.Sprintf("Proposal{%v/%v %v (%v,%v) %v}", p.Height, p.Round,
		p.BlockPartsHeader, p.POLRound, p.POLBlockID, p.Signature)
}

func (p *Proposal) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteJSON(CanonicalJSONOnceProposal{
		ChainID:  chainID,
		Proposal: CanonicalProposal(p),
	}, w, n, err)
}
