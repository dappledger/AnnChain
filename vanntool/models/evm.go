package models

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
)

type CallfJson struct {
	Method   string        `form:"method" json:"function"`
	Params   []interface{} `form:"params" json:"params"`
	ByteCode string        `form:"bytecode" json:"bytecode"`
	Contract string        `form:"contract" json:"contract"`
}

func (cj *CallfJson) JsonString() string {
	str, _ := json.Marshal(cj)
	return string(str)
}

type CreateContract struct {
	Base
	Privkey  string `form:"privkey"`
	Nonce    uint   `form:"nonce"`
	AbiF     string `form:"abif"`
	CallJson string `form:"callf"`
}

func (ec *CreateContract) Args() []string {
	return ParseArgs(ec, append(ec.BaseArgs(), []string{"evm", "create"}...))
}

type CreateContractFull struct {
	CreateContract
	AbiInput string `form:"abi_definition_text"`
	CodeText string `form:"code_text"`
	Params   string `form:"params"`

	CallfJson
}

func (ec *CreateContractFull) FillData(c *beego.Controller) error {
	c.ParseForm(ec)
	ec.BackEnd = parseIPAddr(ec.BackEnd)
	if len(ec.BackEnd) == 0 {
		return fmt.Errorf("backend nil")
	}
	fromAddr, err := PrivToAddress(ec.Privkey)
	if err != nil {
		return err
	}
	ec.Nonce, err = QueryNonce(ec.Base, fromAddr)
	if err != nil {
		return err
	}
	if len(ec.AbiInput) == 0 {
		return fmt.Errorf("abi input shouldn't be null")
	}

	ec.AbiF = ParseStringArg(ec.AbiInput)
	ec.ByteCode = ec.CodeText

	if len(ec.Params) == 0 {
		ec.Params = "[]"
	}
	if err = json.Unmarshal([]byte(ec.Params), &ec.CallfJson.Params); err != nil {
		return err
	}
	ec.CallJson = ec.JsonString()
	return nil
}

func (ec *CreateContractFull) Do() string {
	return ServeCmd(&ec.CreateContract)
}

type CallOrReadContract struct {
	Base
	Nonce    uint   `form:"nonce"`
	AbiF     string `form:"abif"`
	Privkey  string `form:"privkey"`
	CallJson string `form:"callf"`

	Op string
}

func (ec *CallOrReadContract) Args() []string {
	if ec.Op == "read" {
		return ParseArgs(ec, append(ec.BaseArgs(), []string{"evm", "read"}...))
	}
	return ParseArgs(ec, append(ec.BaseArgs(), []string{"evm", "execute"}...))
}

type CallOrReadContractFull struct {
	CallOrReadContract
	AbiInput string `form:"abi_definition_text"`
	Params   string `form:"params"`
	CallfJson
}

func (ec *CallOrReadContractFull) FillData(c *beego.Controller) error {
	c.ParseForm(ec)
	ec.BackEnd = parseIPAddr(ec.BackEnd)
	if len(ec.BackEnd) == 0 {
		return fmt.Errorf("backend nil")
	}
	fromAddr, err := PrivToAddress(ec.Privkey)
	if err != nil {
		return err
	}
	ec.Nonce, err = QueryNonce(ec.Base, fromAddr)
	if err != nil {
		return err
	}
	if len(ec.AbiInput) == 0 {
		return fmt.Errorf("abi input shouldn't be null")
	}
	ec.AbiF = ParseStringArg(ec.AbiInput)
	if err = json.Unmarshal([]byte(ec.Params), &ec.CallfJson.Params); err != nil {
		return err
	}
	ec.CallJson = ec.JsonString()
	return nil
}

func (ec *CallOrReadContractFull) Do() string {
	return ServeCmd(&ec.CallOrReadContract)
}
