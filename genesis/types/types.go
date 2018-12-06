package types

import (
	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

const AccDataLength = 1000

// TypeiUndefined when typei undefined, use 127
const TypeiUndefined = 127

type ShowAccount struct {
	Address string              `json:"address"`
	Balance string              `json:"balance"`
	Data    []map[string]string `json:"data"`
}

type QueryExRequest struct {
	Data   []byte
	Cursor uint64
	Limit  uint
	Order  string
}

type QueryTxRequest struct {
	TxHash *ethcmn.Hash
	Cursor uint64
	Limit  uint
	Order  string
}

type QueryActionsRequest struct {
	Cursor uint64
	Limit  uint
	Order  string
}

type QueryContractExist struct {
	ByteCode string `json:"byte_code"`
	CodeHash string `json:"code_hash"`
	IsExist  bool   `json:"is_exist"`
}
