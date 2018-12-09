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


package ikhofi

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/crypto/sha3"
	"github.com/golang/protobuf/proto"
)

// Sender returns the address derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// Sender may cache the address, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func Sender(signer Signer, tx *Transaction) (common.Address, error) {
	pubkey, err := signer.PublicKey(tx)
	if err != nil {
		return common.Address{}, err
	}
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pubkey[1:])[12:])
	return addr, nil
}

type Signer interface {
	// Hash returns the rlp encoded hash for signatures
	Hash(tx *Transaction) common.Hash
	// PubilcKey returns the public key derived from the signature
	PublicKey(tx *Transaction) ([]byte, error)
	// WithSignature returns a copy of the transaction with the given signature.
	// The signature must be encoded in [R || S || V] format where V is 0 or 1.
	WithSignature(tx *Transaction, sig []byte) (*Transaction, error)
	// Checks for equality on the signers
	Equal(Signer) bool
}

// HomesteadTransaction implements TransactionInterface using the
// homestead rules.
type DawnSigner struct {
}

func (s DawnSigner) Equal(s2 Signer) bool {
	_, ok := s2.(DawnSigner)
	return ok
}

// WithSignature returns a new transaction with the given signature. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (ds DawnSigner) WithSignature(tx *Transaction, sig []byte) (*Transaction, error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for snature: got %d, want 65", len(sig)))
	}
	cpy := &Transaction{
		From:     tx.From,
		To:       tx.To,
		Method:   tx.Method,
		Args:     tx.Args,
		ByteCode: tx.ByteCode,
		Nonce:    tx.Nonce,
	}
	cpy.R = new(big.Int).SetBytes(sig[:32])
	cpy.S = new(big.Int).SetBytes(sig[32:64])
	cpy.V = new(big.Int).SetBytes([]byte{sig[64] + 27})
	cpy.Hash = ds.Hash(cpy)
	return cpy, nil
}

func (ds DawnSigner) PublicKey(tx *Transaction) ([]byte, error) {
	if tx.V.BitLen() > 8 {
		return nil, ErrInvalidSig
	}
	V := byte(tx.V.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, tx.R, tx.S, true) {
		return nil, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := tx.R.Bytes(), tx.S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V

	// recover the public key from the snature
	hash := ds.Hash(tx)
	pub, err := crypto.Ecrecover(hash[:], sig)
	if err != nil {
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}
	return pub, nil
}

func GetHash(tmp *TransactionPbTmp) (h common.Hash) {
	hw := sha3.NewKeccak256()
	bs, _ := proto.Marshal(tmp)
	hw.Write(bs)
	hw.Sum(h[:0])
	return h
}

// Hash returns the hash to be sned by the sender.
// It does not uniquely identify the transaction.
func (ds DawnSigner) Hash(tx *Transaction) common.Hash {
	tmp := tx.Transaction2PbTmp()
	return GetHash(tmp)
}
