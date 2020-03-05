package crypto

import (
	pbcrypto "github.com/dappledger/AnnChain/gemmill/protos/crypto"
)

func (p SignatureEd25519) ToPbData() *pbcrypto.Signature {
	var pk pbcrypto.Signature
	pk.Type = pbcrypto.KeyType_Ed25519
	pk.Bytes = p.Bytes()
	return &pk
}

func (p SignatureSecp256k1) ToPbData() *pbcrypto.Signature {
	var pk pbcrypto.Signature
	pk.Type = pbcrypto.KeyType_Secp256k1
	pk.Bytes = p.Bytes()
	return &pk
}

func (p PubKeyEd25519) ToPbData() *pbcrypto.PubKey {
	var pk pbcrypto.PubKey
	pk.Type = pbcrypto.KeyType_Ed25519
	pk.Bytes = p.Bytes()
	return &pk
}

func (p PubKeySecp256k1) ToPbData() *pbcrypto.PubKey {
	var pk pbcrypto.PubKey

	pk.Type = pbcrypto.KeyType_Secp256k1
	pk.Bytes = p.Bytes()
	return &pk
}
