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
	"fmt"
	"os"

	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/vanntool/tools"
)

type JvmContract struct {
	Base
	Privkey    string `form:"privkey"`
	Method     string `form:"method"`
	ContractID string `form:"contractid"`
	ByteCode   string `form:"bytecode"`
}

func (ec *JvmContract) Args() []string {
	return ParseArgs(ec, append(ec.BaseArgs(), []string{"ikhofi", "execute"}...))
}

type CreateJvmContractFull struct {
	JvmContract

	filePath string
}

func (ec *CreateJvmContractFull) FillData(c *beego.Controller) error {
	c.ParseForm(ec)
	ec.BackEnd = parseIPAddr(ec.BackEnd)
	if len(ec.BackEnd) == 0 {
		return fmt.Errorf("backend nil")
	}
	var fileName string
	var err error
	fileName, ec.filePath, err = LoadAndSaveTempFile(c, "file")
	ec.ByteCode = ParseStringArg(ec.filePath)
	ec.Method = fmt.Sprintf("deploy('%v','%v')", ec.ContractID, fileName)
	ec.ContractID = "system"
	return err
}

func (ec *CreateJvmContractFull) Do() string {
	ret := ServeCmd(&ec.JvmContract)
	os.Remove(ec.filePath)
	return ret
}

type CallOrQueryJvmContract struct {
	Base
	Privkey    string `form:"privkey"`
	Method     string `form:"method"`
	ContractID string `form:"contractid"`
	Op         string
}

func (ec *CallOrQueryJvmContract) Args() []string {
	if ec.Op == "call" {
		return ParseArgs(ec, append(ec.BaseArgs(), []string{"ikhofi", "execute"}...))
	}
	return ParseArgs(ec, append(ec.BaseArgs(), []string{"ikhofi", "query"}...))
}

type CallOrQueryJvmContractFull struct {
	CallOrQueryJvmContract
	Params string `form:"params"`
}

func (cq *CallOrQueryJvmContractFull) FillData(c *beego.Controller) error {
	c.ParseForm(cq)
	cq.BackEnd = parseIPAddr(cq.BackEnd)
	if len(cq.BackEnd) == 0 {
		return fmt.Errorf("backend nil")
	}
	if !tools.OnlyNumLetterUnderline(cq.Method) {
		return fmt.Errorf("method name syntax error")
	}
	cq.Method = fmt.Sprintf("%v(%v)", cq.Method, cq.Params)
	return nil
}

func (ec *CallOrQueryJvmContractFull) Do() string {
	return ServeCmd(&ec.CallOrQueryJvmContract)
}
