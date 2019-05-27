package types

import (
	crypto "github.com/dappledger/AnnChain/gemmill/go-crypto"
)

func SignCA(priv crypto.PrivKey, pubbytes []byte) string {
	return priv.Sign(pubbytes).KeyString()
}
