package tools

import (
	"github.com/dappledger/AnnChain/eth/crypto/sha3"
)

func HashKeccak(data []byte) ([]byte, error) {
	keccak := sha3.NewKeccak256()
	if _, err := keccak.Write(data); err != nil {
		return nil, err
	}
	return keccak.Sum(nil), nil
}
