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

	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
)

//------------------------------------------------------------
// we store the gendoc in the db

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

type GenesisValidator struct {
	PubKey crypto.StPubKey `json:"pub_key"`
	Amount int64           `json:"amount"`
	Name   string          `json:"name"`
	IsCA   bool            `json:"is_ca"`
}

func (gv *GenesisValidator) UnmarshalJSON(data []byte) error {
	st := struct {
		PubKey crypto.StPubKey `json:"pub_key"`
		Amount int64           `json:"amount"`
		Name   string          `json:"name"`
		IsCA   bool            `json:"is_ca"`
	}{}
	err := json.Unmarshal(data, &st)
	if err != nil {
		return err
	}
	gv.PubKey = st.PubKey
	gv.Amount = st.Amount
	gv.Name = st.Name
	gv.IsCA = st.IsCA
	return nil
}

type GenesisDoc struct {
	GenesisTime Time               `json:"genesis_time"`
	ChainID     string             `json:"chain_id"`
	Validators  []GenesisValidator `json:"validators"`
	AppHash     Bytes              `json:"app_hash"`
	Plugins     string             `json:"plugins"`
}

// Utility method for saving GenensisDoc as JSON file.
func (genDoc *GenesisDoc) SaveAs(file string) error {
	genDocBytes, err := json.MarshalIndent(genDoc, "", "\t")
	if err != nil {
		return err
	}
	return WriteFile(file, genDocBytes, 0644)
}

func (genDoc *GenesisDoc) JSONBytes() ([]byte, error) {
	return json.Marshal(genDoc)
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (genState *GenesisDoc) {
	genState = &GenesisDoc{}
	err := json.Unmarshal(jsonBlob, genState)
	if err != nil {
		Exit(Fmt("Couldn't read GenesisDoc: %v", err))
	}
	return
}

func GenesisDocFromJSONRet(jsonBlob []byte) (genState *GenesisDoc, err error) {
	genState = &GenesisDoc{}
	err = json.Unmarshal(jsonBlob, genState)
	return
}

type GenesisValidatorJson struct {
	PubKey [32]byte `json:"pub_key"`
	Amount int64    `json:"amount"`
	Name   string   `json:"name"`
	IsCA   bool     `json:"is_ca"`
}

func (gv *GenesisValidatorJson) UnmarshalJSON(b []byte) error {
	gj := GenesisValidatorJson{}
	if err := json.Unmarshal(b, &gj); err != nil {
		return err
	}
	gv.Amount = gj.Amount
	gv.IsCA = gj.IsCA
	gv.Name = gj.Name
	pk := crypto.PubKeyEd25519(gj.PubKey)
	gv.PubKey = pk
	return nil
}
