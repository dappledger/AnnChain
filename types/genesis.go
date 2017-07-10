package types

import (
	"time"

	"encoding/json"
	. "gitlab.zhonganonline.com/ann/ann-module/lib/go-common"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-crypto"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-wire"
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
	return WriteFile(file, genDocBytes, 0644)
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (genState *GenesisDoc) {
	var err error
	wire.ReadJSONPtr(&genState, jsonBlob, &err)
	if err != nil {
		Exit(Fmt("Couldn't read GenesisDoc: %v", err))
	}
	return
}

type GenesisValidatorJson struct {
	PubKey     [32]byte `json:"pub_key"`
	Amount     int64    `json:"amount"`
	Name       string   `json:"name"`
	IsCA       bool     `json:"is_ca"`
	RPCAddress string   `json:"rpc"`
}

func (gv *GenesisValidator) UnmarshalJSON(b []byte) error {
	gj := GenesisValidatorJson{}
	if err := json.Unmarshal(b, &gj); err != nil {
		return err
	}
	gv.Amount = gj.Amount
	gv.IsCA = gj.IsCA
	gv.Name = gj.Name
	gv.PubKey = crypto.PubKeyEd25519(gj.PubKey)
	gv.RPCAddress = gj.RPCAddress
	return nil
}
