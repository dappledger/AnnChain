package types

import (
	"encoding/json"
	"time"
)

type QueryTxInfo struct {
	Height        int       `json:"height"`
	BlockHash     []byte    `json:"blockhash"`
	BlockTime     time.Time `json:"blocktime"`
	ValidatorHash []byte    `json:"validatorhash"`
}

func (i *QueryTxInfo) ToBytes() ([]byte, error) {
	return json.Marshal(i)
}

func (i *QueryTxInfo) FromBytes(bytes []byte) error {
	return json.Unmarshal(bytes, i)
}
