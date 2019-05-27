// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"encoding/json"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
)

//------------------------------------------------------------
// we store the gendoc in the db

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

type GenesisValidator struct {
	PubKey     crypto.PubKey `json:"pub_key"`
	Amount     int64         `json:"amount"`
	Name       string        `json:"name"`
	IsCA       bool          `json:"is_ca"`
	RPCAddress string        `json:"rpc"`
}

type GenesisDoc struct {
	GenesisTime time.Time          `json:"genesis_time"`
	ChainID     string             `json:"chain_id"`
	Validators  []GenesisValidator `json:"validators"`
	AppHash     []byte             `json:"app_hash"`
	Plugins     string             `json:"plugins"`
}

// Utility method for saving GenensisDoc as JSON file.
func (genDoc *GenesisDoc) SaveAs(file string) error {
	genDocBytes := wire.JSONBytesPretty(genDoc)
	return gcmn.WriteFile(file, genDocBytes, 0644)
}

func (genDoc *GenesisDoc) JSONBytes() []byte {
	return wire.JSONBytesPretty(genDoc)
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) *GenesisDoc {
	genState, err := GenesisDocFromJSONRet(jsonBlob)
	if err != nil {
		gcmn.Exit(gcmn.Fmt("Couldn't read GenesisDoc: %v", err))
	}
	return genState
}

func GenesisDocFromJSONRet(jsonBlob []byte) (genState *GenesisDoc, err error) {
	wire.ReadJSONPtr(&genState, jsonBlob, &err)
	return
}

type GenesisValidatorJson struct {
	PubKey     []byte `json:"pub_key"`
	Amount     int64  `json:"amount"`
	Name       string `json:"name"`
	IsCA       bool   `json:"is_ca"`
	RPCAddress string `json:"rpc"`
}

func (gv *GenesisValidator) UnmarshalJSON(b []byte) error {
	var err error
	gj := GenesisValidatorJson{}
	if err := json.Unmarshal(b, &gj); err != nil {
		return err
	}
	gv.Amount = gj.Amount
	gv.IsCA = gj.IsCA
	gv.Name = gj.Name
	gv.PubKey, err = crypto.PubKeyFromBytes(gj.PubKey)
	if err != nil {
		return err
	}
	gv.RPCAddress = gj.RPCAddress
	return nil
}
