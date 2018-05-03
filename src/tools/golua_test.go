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
