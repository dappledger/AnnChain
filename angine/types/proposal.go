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

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

var (
	ErrInvalidBlockPartSignature = errors.New("Error invalid block part signature")
	ErrInvalidBlockPartHash      = errors.New("Error invalid block part hash")
)

// polRound: -1 if no polRound.
func NewProposal(height, round def.INT, blockPartsHeader pbtypes.PartSetHeader, polRound def.INT, polBlockID pbtypes.BlockID) *pbtypes.Proposal {
	return &pbtypes.Proposal{
		Data: &pbtypes.ProposalData{
			Height:           height,
			Round:            round,
			BlockPartsHeader: &blockPartsHeader,
			POLRound:         polRound,
			POLBlockID:       &polBlockID,
		},
	}
}
