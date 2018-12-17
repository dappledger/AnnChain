// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/crypto"
)

// Signer de-sign from signature and get signer's address(publickey)
func Signer(tx *Transaction, sig []byte) (ethcmn.Address, error) {
	if len(sig) != 65 {
		return ethcmn.Address{}, errors.New("invalid signature length")
	}

	sigHash := tx.SigHash()

	publicKey, err := crypto.Ecrecover(sigHash[:], sig)
	if err != nil {
		return ethcmn.Address{}, err
	}
	if len(publicKey) == 0 || publicKey[0] != 4 {
		return ethcmn.Address{}, errors.New("invalid public key")
	}
	return ethcmn.BytesToAddress(crypto.Keccak256(publicKey[1:])[12:]), nil
}

// Sign tx using signatures
func (tx *Transaction) Sign(privkey *ecdsa.PrivateKey) (*Transaction, error) {

	sigHash := tx.SigHash()

	sig, err := crypto.Sign(sigHash.Bytes(), privkey)
	if err != nil {
		return nil, err
	}

	tx.Data.Sign = hex.EncodeToString(sig)

	return tx, nil
}

// CheckSig check signature
func (tx *Transaction) CheckSig() error {
	if len(tx.SignString()) == 0 {
		return fmt.Errorf("no signature")
	}
	signer, err := Signer(tx, tx.Signature())
	if err != nil {
		return err
	}
	if 0 != signer.Compare(tx.GetFrom()) {
		return fmt.Errorf("signature not from address")
	}
	return nil
}
