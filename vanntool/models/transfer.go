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


package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/eth/crypto"
)

func PrivToAddress(privStr string) (string, error) {
	if strings.HasPrefix(privStr, "0x") {
		privStr = privStr[2:]
	}
	privkey, err := crypto.HexToECDSA(privStr)
	if err != nil {
		return "", err
	}
	addr := crypto.PubkeyToAddress(privkey.PublicKey)
	return addr.Hex(), nil
}

func QueryNonce(base Base, addr string) (uint, error) {
	var query QueryFull
	query.Base = base
	query.Addr = addr
	return query.Nonce()
}

func QueryBalance(base Base, addr string) (uint, error) {
	var query QueryFull
	query.Base = base
	query.Addr = addr
	return query.Balance()
}

//////////////////////////////////////////////////////////////////////////////////

type Transfer struct {
	Base
	Privkey string `form:"privkey"`
	To      string `form:"to"`
	Nonce   uint   `form:"nonce"`
	Value   uint   `form:"value"`
	parsed  bool
}

func (t *Transfer) Args() []string {
	t.BackEnd = parseIPAddr(t.BackEnd)
	return ParseArgs(t, append(t.BaseArgs(), []string{"tx", "send"}...))
}

type TransferFull struct {
	Transfer
}

func (f *TransferFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	fromAddr, err := PrivToAddress(f.Privkey)
	if err != nil || len(fromAddr) == 0 {
		return fmt.Errorf("decrypt to addr error:", err)
	}
	f.Nonce, err = QueryNonce(f.Base, fromAddr)
	return err
}

func (f *TransferFull) Do() string {
	return ServeCmd(f)
}

//////////////////////////////////////////////////////////////////////////////////

type QueryTx struct {
	Base
	Receipt string `form:"hash"`
}

func (q *QueryTx) Args() []string {
	q.BackEnd = parseIPAddr(q.BackEnd)
	return ParseArgs(q, append(q.BaseArgs(), []string{"query", "receipt"}...))
}

type QueryTxFull struct {
	QueryTx
}

func (f *QueryTxFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	return nil
}

func (f *QueryTxFull) Do() string {
	return ServeCmd(f)
}

//////////////////////////////////////////////////////////////////////////////////

type Query struct {
	Base
	Addr string `form:"address"`
}

func (q *Query) args(qcmd string) []string {
	q.BackEnd = parseIPAddr(q.BackEnd)
	return ParseArgs(q, append(q.BaseArgs(), []string{"query", qcmd}...))
}

type QueryFull struct {
	Query
	Qcmd string `form:"querycmd"`
}

func (f *QueryFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	return nil
}

func (f *QueryFull) Args() []string {
	return f.args(f.Qcmd)
}

func (f *QueryFull) Do() string {
	var ai AccountInfo
	var err error
	ai.Nonce, err = f.Nonce()
	if err != nil {
		return err.Error()
	}
	ai.Balance, err = f.Balance()
	if err != nil {
		return err.Error()
	}
	return ai.ToString()
}

func (f *QueryFull) doShell() string {
	return ServeCmd(f)
}

func _regexNum(str string) (uint, error) {
	nonce := strings.Trim(str, " \n")
	index := strings.Index(nonce, ":")
	if index < 0 {
		return 0, fmt.Errorf(nonce)
	}
	nonce = strings.Trim(nonce[index+1:], " \n")
	i, err := strconv.Atoi(nonce)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

func (f *QueryFull) Nonce() (uint, error) {
	f.Qcmd = "nonce"
	return _regexNum(f.doShell())
}

func (f *QueryFull) Balance() (uint, error) {
	f.Qcmd = "balance"
	return _regexNum(f.doShell())
}

type AccountInfo struct {
	Nonce   uint `json:"nonce"`
	Balance uint `json:"balance"`
}

func (ai *AccountInfo) ToString() string {
	by, err := json.Marshal(ai)
	if err != nil {
		return ""
	}
	return string(by)
}
