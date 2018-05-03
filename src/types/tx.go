package types

import (
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
)

type ICivilTx interface {
	GetPubKey() []byte
	SetPubKey([]byte)
	PopSignature() []byte
	SetSignature([]byte)
	Sender() []byte
}

type CivilTx struct {
	sender []byte

	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`
}

func (t *CivilTx) GetPubKey() []byte {
	return t.PubKey
}

func (t *CivilTx) SetPubKey(pk []byte) {
	t.PubKey = pk
}

func (t *CivilTx) PopSignature() []byte {
	s := t.Signature
	t.Signature = nil
	return s
}

func (t *CivilTx) SetSignature(s []byte) {
	t.Signature = s
}

func (t *CivilTx) Sender() []byte {
	if t.sender != nil || len(t.sender) > 0 {
		return t.sender
	}

	pubkey := crypto.PubKeyEd25519{}
	copy(pubkey[:], t.PubKey)
	t.sender = pubkey.Address()

	return t.sender
}
