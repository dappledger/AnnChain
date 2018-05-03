package types

import (
	"bytes"
	"encoding/json"
)

var (
	SuspectTxTag = []byte{'s', 'p', 't', 0x01}
)

type SuspectTx struct {
	Suspect   *Hypocrite `json:"suspect"`
	PubKey    []byte     `json:"pubkey"`
	Signature []byte     `json:"signature"`
}

func IsSuspectTx(tx []byte) bool {
	return bytes.Equal(SuspectTxTag, tx[:4])
}

func (tx *SuspectTx) ToBytes() ([]byte, error) {
	return json.Marshal(tx)
}

func (tx *SuspectTx) FromBytes(bs []byte) error {
	return json.Unmarshal(bs, tx)
}
