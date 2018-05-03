/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package tools

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	crypto "github.com/dappledger/AnnChain/module/lib/go-crypto"
	cvtypes "github.com/dappledger/AnnChain/src/types"
)

// TxToBytes defines a universal way to serialize ICivilTx
func TxToBytes(tx cvtypes.ICivilTx) ([]byte, error) {
	return json.Marshal(tx)
}

// TxFromBytes is just the reverse operation of TxToBytes
func TxFromBytes(bytes []byte, tx cvtypes.ICivilTx) error {
	tType, tValue := reflect.TypeOf(tx), reflect.ValueOf(tx)
	if tType.Kind() != reflect.Ptr {
		return errors.Errorf("the second param should be a pointer")
	}
	if tValue.IsNil() {
		return errors.Errorf("the second param should not be nil")
	}

	return json.Unmarshal(bytes, tx)
}

// TxSign now will also embed the pubkey of the signer, don't need filling pubkey manually
func TxSign(tx cvtypes.ICivilTx, p crypto.PrivKey) ([]byte, error) {
	priv, ok := p.(*crypto.PrivKeyEd25519)
	if !ok {
		return nil, errors.Wrap(errors.Errorf("private key fails, only support Ed25519"), "[TxSign]")
	}
	pubkey := priv.PubKey().(*crypto.PubKeyEd25519)
	tx.SetPubKey(pubkey[:])
	txBytes, err := TxToBytes(tx)
	if err != nil {
		return nil, errors.Wrap(err, "[TxSign]")
	}
	sig := priv.Sign(txBytes).(*crypto.SignatureEd25519)
	tx.SetSignature(sig[:])
	return sig[:], nil
}

// TxHash hashes ICivilTx
func TxHash(tx cvtypes.ICivilTx) ([]byte, error) {
	txBytes, err := TxToBytes(tx)
	if err != nil {
		return nil, err
	}
	return HashKeccak(txBytes)
}

// TxVerifySignature verifies the signature carried by the ICivilTx
func TxVerifySignature(tx cvtypes.ICivilTx) (bool, error) {
	sig := tx.PopSignature()
	defer tx.SetSignature(sig)

	pubkey := crypto.PubKeyEd25519{}
	copy(pubkey[:], tx.GetPubKey())
	signature := crypto.SignatureEd25519{}
	copy(signature[:], sig)
	tBytes, err := TxToBytes(tx)
	if err != nil {
		return false, err
	}
	s64 := [64]byte(signature)
	p32 := [32]byte(pubkey)
	return ed25519.Verify(&p32, tBytes, &s64), nil
}
