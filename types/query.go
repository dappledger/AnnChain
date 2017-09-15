package types

import (
	"encoding/json"
	"time"
)

const (
	QueryTxExecution = 0x01
)

type TxExecutionResult struct {
	Height        int       `json:"height"`
	BlockHash     []byte    `json:"blockhash"`
	BlockTime     time.Time `json:"blocktime"`
	ValidatorHash []byte    `json:"validatorhash"`
}

func (i *TxExecutionResult) ToBytes() ([]byte, error) {
	return json.Marshal(i)
}

func (i *TxExecutionResult) FromBytes(bytes []byte) error {
	return json.Unmarshal(bytes, i)
}
