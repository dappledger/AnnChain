package consensus

import "github.com/dappledger/AnnChain/gemmill/types"

type Engine interface {
	GetValidators() (int64, []*types.Validator)
	SetEventSwitch(types.EventSwitch)
	ValidateBlock(*types.Block) error
}
