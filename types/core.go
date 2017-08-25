package types

import (
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-crypto"
)

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
}
