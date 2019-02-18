// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"sync/atomic"
	"time"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/rlp"
)

var ErrInvalidSig = errors.New("invalid transaction v, r, s values")

var (
	errMissingTxSignatureFields = errors.New("missing required JSON transaction signature fields")
	errMissingTxFields          = errors.New("missing required JSON transaction fields")
	errNoSigner                 = errors.New("missing signing methods")

	maxTxSigs   = 20
	nForBigdata = big.NewInt(5)
)

type Transaction struct {
	Data TxData
	// caches
	creatime uint64
	option   OperatorItfc
	hash     atomic.Value
	size     atomic.Value
	from     atomic.Value
}

type TxData struct {
	Basefee   string
	Nonce     string
	Memo      string
	OpType    string
	From      string
	To        string
	Operation string
	Sign      string
}

func (tx *Transaction) SetOperatorItfc(op OperatorItfc) {
	tx.option = op
}

func (tx *Transaction) GetOperatorItfc() OperatorItfc {
	return tx.option
}

func (tx *Transaction) String() string {
	return fmt.Sprintf("nonce:%v,from:%v,to:%v,basefee:%v,optype:%v,operation:%v,memo:%v,createat:%v,sign:%v",
		tx.Data.Nonce, tx.Data.From, tx.Data.To,
		tx.Data.Basefee, tx.Data.OpType, tx.Data.Operation, tx.Data.Memo, tx.GetCreateTime(), tx.Data.Sign)
}

func (tx *Transaction) SignString() string {
	return tx.Data.Sign
}

func (tx *Transaction) Signature() []byte {

	bSign, _ := ethcmn.HexToByte(tx.SignString())

	return bSign
}

func (tx *Transaction) SetFrom(from string) {
	tx.Data.From = from
}

func (tx *Transaction) GetFrom() ethcmn.Address {
	return ethcmn.HexToAddress(tx.Data.From)
}

func (tx *Transaction) GetOpName() string {
	return tx.Data.OpType
}

func (tx *Transaction) GetOperation() []byte {
	return []byte(tx.Data.Operation)
}

func (tx *Transaction) Decode2Operation(st interface{}) error {
	return rlp.DecodeBytes(tx.GetOperation(), st)
}

func NewTransaction(nonce, basefee, from, to, optype, memo, op string) *Transaction {
	d := TxData{
		Nonce:     nonce,
		Basefee:   basefee,
		From:      from,
		To:        to,
		OpType:    optype,
		Memo:      memo,
		Operation: op,
	}
	return &Transaction{Data: d}
}

// DecodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.Data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.Data)
	if err == nil {
		tx.size.Store(ethcmn.StorageSize(rlp.ListSize(size)))
	}
	return err
}

func (tx *Transaction) SetCreateTime(t uint64) {
	tx.creatime = t
}

func (tx *Transaction) GetCreateTime() uint64 {
	return tx.creatime
}

func (tx *Transaction) BaseFee() *big.Int {
	bigFee, _ := new(big.Int).SetString(tx.Data.Basefee, 10)
	return bigFee
}

func (tx *Transaction) Nonce() uint64 {
	nonce, err := strconv.ParseUint(tx.Data.Nonce, 10, 64)
	if err != nil {
		return 0
	}
	return nonce
}

func (tx *Transaction) CheckNonce() bool { return true }

func (tx *Transaction) GetTo() ethcmn.Address {
	return ethcmn.HexToAddress(tx.Data.To)
}

func (tx *Transaction) Hash() ethcmn.Hash {
	return tx.SigHash()
}

func (tx *Transaction) SigHash() ethcmn.Hash {

	if hash := tx.hash.Load(); hash != nil {
		return hash.(ethcmn.Hash)
	}
	v := rlpHash([]interface{}{
		tx.Data.Basefee,
		tx.Data.Nonce,
		tx.Data.Memo,
		tx.Data.OpType,
		tx.Data.From,
		tx.Data.To,
		tx.Data.Operation,
	})

	tx.hash.Store(v)
	return v
}

// Transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Data.Nonce < s[j].Data.Nonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (tx *Transaction) GetDBTxData(exeRes error) *TransactionData {
	d := &TransactionData{
		TxHash:          tx.SigHash(),
		CreateDate:      time.Now().UnixNano(),
		Account:         tx.GetFrom(),
		Target:          tx.GetTo(),
		OpType:          tx.GetOpName(),
		AccountSequence: new(big.Int).SetUint64(tx.Nonce()),
		FeePaid:         tx.BaseFee(),
		Memo:            string(tx.Data.Memo),
	}

	if exeRes != nil {
		d.ResultCode = 500 // int -> uint
		d.ResultCodes = exeRes.Error()
	} else {
		d.ResultCode = 0
		d.ResultCodes = "success"
	}

	return d
}

type Receipt struct {
	Nonce           uint64             `json:"nonce"`
	TxHash          ethcmn.Hash        `json:"hash"`
	TxReceiptStatus bool               `json:"tx_receipt_status"`
	Message         string             `json:"msg"`
	ContractAddress string             `json:"contract_address"`
	GasUsed         *big.Int           `json:"gas_used"`
	Height          uint64             `json:"height"`
	Source          ethcmn.Address     `json:"from"`
	Payload         string             `json:"payload"`
	GasPrice        string             `json:"gas_price"`
	GasLimit        string             `json:"gas_limit"`
	Logs            []*json.RawMessage `json:"logs"`
	OpType          OP_NAME            `json:"optype"`
	Res             string             `json:"result"`
}

func (recpt *Receipt) String() string {
	return fmt.Sprintf("nonce:%v,TxHash:%v,txReceiptStatus:%v,Message:%v,ContractAddress:%v,GasUsed:%v,Height:%v,Source:%v,Payload:%v,GasPrice:%v,GasLimit:%v,Logs:%v,OpType:%v,Res:%v",
		recpt.Nonce, recpt.TxHash.Hex(), recpt.TxReceiptStatus,
		recpt.Message, recpt.ContractAddress, recpt.GasUsed, recpt.Height, recpt.Source.Hex(),
		recpt.Payload, recpt.GasPrice, recpt.GasLimit, recpt.Logs, recpt.OpType, recpt.Res)

}
