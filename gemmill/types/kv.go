package types

import (
	"encoding/hex"
	"fmt"
)

type ValueUpdateHistory struct {
	TxHash      []byte `json:"tx_hash"`
	BlockHeight uint64 `json:"block_height"`
	TimeStamp   uint64 `json:"time_stamp"`
	Value       []byte `json:"value"`
	TxIndex     uint32  `json:"tx_index"`
}

type ValueHistoryResult struct {
	Key                  []byte                `json:"key"`
	ValueUpdateHistories []*ValueUpdateHistory `json:"value_update_histories"`
	Total                uint32                `json:"total"`
}

func (v *ValueUpdateHistory) String() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("hight : %d , hash :%s ,time: %d , value %s , txIndx %d",
		v.BlockHeight, hex.EncodeToString(v.TxHash), v.TimeStamp, string(v.Value),v.TxIndex)
}

type KeyValueHistory struct {
	Key                []byte
	ValueUpdateHistory *ValueUpdateHistory
}

func (k *KeyValueHistory)String()string {
	if k == nil {
		return ""
	}
	return fmt.Sprintf("key : %s ,history: %s",string(k.Key),k.ValueUpdateHistory.String())
}

type KeyValueHistories []*KeyValueHistory

func (k KeyValueHistories)String()string {
	if len(k) ==0 {
		return ""
	}
	result := "["
	for _, v := range k {
		result += v.String()+" "
	}
	result+="]"
	return result
}