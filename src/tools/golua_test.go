package tools

import (
	"testing"
)

const (
	codeOrgA1 = `
	function main(params)
		if params["contract_call"] == nil or params.contract_call["function"] ~= "buyChicken" then 
			return nil
		end
		return {                                                                                
			["from"] = params.from,
			["to"] = params.to,
			["value"] = params.value,
			["nonce"] = params.nonce,
			["function"] = params.contract_call["function"],
			["score"] = params.contract_call["_score"] 
		};
	end`
	codeOrgB1 = `
	function main(params)
		if tonumber(params.score) < 1000 then
			return nil
		else
			params["params"] = params.score
			params["contractid"] = "Chicken"
			params["ikhofi_func"] = "buyChickenByScores"
			return params
		end
	end`
)

func TestGoLua2(t *testing.T) {
	L := NewLuaState()
	outParams := make(map[string]interface{})
	contractParams := make(map[string]interface{})
	contractParams["function"] = "buyChicken"
	contractParams["_score"] = 1888
	outParams["contract_call"] = contractParams
	outParams["from"] = "aa"
	outParams["to"] = "bb"
	outParams["value"] = 666
	outParams["nonce"] = 1
	ret, err := ExecLuaWithParam(L, codeOrgA1, outParams)
	if err != nil {
		t.Error("codeOrgA1", err)
	}
	L.Close()

	L = NewLuaState()
	ret, err = ExecLuaWithParam(L, codeOrgB1, ret)
	if err != nil {
		t.Error("codeOrgB1", err, ret)
	}

	if ret["params"].(float64) != 1888.0 {
		t.Error("params return err number", err, ret)
	}
	L.Close()
}
