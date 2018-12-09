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
	"math/big"
	"time"

	"github.com/dappledger/AnnChain/eth/common"
)

type Transaction struct {
	From     common.Address
	To       string
	Method   string
	Args     []string
	ByteCode []byte
	Nonce    int64
	V        *big.Int // signature
	R, S     *big.Int // signature
	Hash     common.Hash
}

var ErrInvalidSig = errors.New("invalid transaction v, r, s values")

func NewTransaction(from common.Address, to, method string, args []string, bytecode []byte) *Transaction {
	return newTransaction(from, to, method, args, bytecode)
}

func newTransaction(from common.Address, to, method string, args []string, bytecode []byte) *Transaction {
	now := time.Now()
	nonce := now.UTC().UnixNano()

	tx := Transaction{
		From:     from,
		To:       to,
		Method:   method,
		Args:     args,
		ByteCode: bytecode,
		Nonce:    nonce,
	}

	return &tx
}

func (tx *Transaction) Transaction2Pb() *TransactionPb {
	return &TransactionPb{
		From:     tx.From[:],
		To:       tx.To,
		Method:   tx.Method,
		Args:     tx.Args,
		Bytecode: tx.ByteCode,
		Nonce:    tx.Nonce,
		V:        tx.V.Bytes(),
		R:        tx.R.Bytes(),
		S:        tx.S.Bytes(),
		Hash:     tx.Hash[:],
	}
}

func (tx *Transaction) Transaction2PbTmp() *TransactionPbTmp {
	return &TransactionPbTmp{
		From:     tx.From[:],
		To:       tx.To,
		Method:   tx.Method,
		Args:     tx.Args,
		Bytecode: tx.ByteCode,
		Nonce:    tx.Nonce,
	}
}

func Pb2Transaction(txpb *TransactionPb) *Transaction {
	V := new(big.Int)
	R := new(big.Int)
	S := new(big.Int)
	return &Transaction{
		From:     common.BytesToAddress(txpb.From),
		To:       txpb.To,
		Method:   txpb.Method,
		Args:     txpb.Args,
		ByteCode: txpb.Bytecode,
		Nonce:    txpb.Nonce,
		V:        V.SetBytes(txpb.V),
		R:        R.SetBytes(txpb.R),
		S:        S.SetBytes(txpb.S),
		Hash:     common.BytesToHash(txpb.Hash),
	}
}

// SigHash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (tx *Transaction) SigHash(signer Signer) common.Hash {
	return signer.Hash(tx)
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	return signer.WithSignature(tx, sig)
}
