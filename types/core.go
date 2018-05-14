package types

import (
	"gitlab.zhonganinfo.com/tech_bighealth/ann-module/lib/go-crypto"
)

type Superior interface {
	BroadcastTx([]byte) error
}

type Core interface {
	IsValidator() bool
	GetPublicKey() (crypto.PubKeyEd25519, bool)
	GetPrivateKey() (crypto.PrivKeyEd25519, bool)
	GetChainID() string
	GetEngine() Engine
}

type Engine interface {
	GetBlock(int) (*Block, *BlockMeta)
	GetValidators() (int, *ValidatorSet)
	PrivValidator() *PrivValidator
	BroadcastTx([]byte) error
	Query(byte, []byte) (interface{}, error)
}
