package types

import (
	"math/big"
)

var (
	_BIG_INT_0 = big.NewInt(0)
)

const (
	CODE_VAR_ENT = "ent_params"
	CODE_VAR_RET = "ret_params"
)

func BigInt0() *big.Int {
	return _BIG_INT_0
}

type ParamUData = map[string]interface{}

type QueryType = byte

const (
	QueryType_Contract  QueryType = 0
	QueryType_Nonce     QueryType = 1
	QueryType_Balance   QueryType = 2
	QueryType_Receipt   QueryType = 3
	QueryType_Existence QueryType = 4
)
