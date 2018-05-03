package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-wire"
)

type HTTPResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      string           `json:"id"`
	Result  *json.RawMessage `json:"result"`
	Error   string           `json:"error"`
}

func (h *HTTPResponse) json() []byte {
	bytes, _ := json.Marshal(h)
	return bytes
}

func NewHTTPResponse(data agtypes.RPCResult) *HTTPResponse {
	var raw *json.RawMessage
	if data != nil {
		var pdata *agtypes.RPCResult
		pdata = &data
		rawMsg := json.RawMessage(wire.JSONBytes(pdata))
		raw = &rawMsg
	}
	return &HTTPResponse{
		JSONRPC: "2.0",
		Result:  raw,
	}
}

type GenBlockForTest struct {
	beego.Controller
}

func (c *GenBlockForTest) Get() {
	chainid := c.Input().Get("chainid")
	heightStr := c.Input().Get("height")
	height, err := strconv.ParseInt(heightStr, 10, 0)
	if err != nil {
		fmt.Println("[gen block],parse string height to int err:", err)
	}
	block := agtypes.GenBlockForTest(height, 5, 5)
	block.Header.ChainID = chainid
	var result agtypes.ResultBlock
	result.Block = block.Block
	res := NewHTTPResponse(&result)
	c.Ctx.Output.Body(res.json())
}

type LastHeightForTest struct {
	beego.Controller
}

func (c *LastHeightForTest) Get() {
	_ = c.Input().Get("chainid")
	var result agtypes.ResultLastHeight
	result.LastHeight = 100 // XD
	res := NewHTTPResponse(&result)
	c.Ctx.Output.Body(res.json())
}
