package types

import (
	"bytes"
	"time"
)

type SpecialOPCmd struct {
	CmdCode      string    `json:"cmdcode"` //code for check if im a special
	CmdType      string    `json:"cmdtype"` //type for what kind of specialOP
	Msg          []byte    `json:"msg"`
	Sigs         [][]byte  `json:"sigs"`
	ExCmd        Tx        `json:"excmd"`
	NodePubKey   []byte    `json:"node_pubkey"`
	IssuerPubKey []byte    `json:"issuer_pubkey"`
	Time         time.Time `json:"time"`
	Nonce        uint64    `json:"nonce"`
}

type CmdType string

const (
	SpecialOP                  = "specialOP"
	SpecialOP_ChangeValidator  = "changeValidator"
	SpecialOP_Disconnect       = "disconnect"
	SpecialOP_PromoteValidator = "promoteValidator"
	SpecialOP_DeleteValidator  = "deleteValidator"
	SpecialOP_ChangePower      = "changePower"
	SpecialOP_AddRefuseKey     = "addRefuseKey"
	SpecialOP_DeleteRefuseKey  = "deleteRefuseKey"
)

func TagSpecialOPTx(tx []byte) []byte {
	return WrapTx([]byte("zaop"), tx)
}

func IsSpecialOP(tx []byte) bool {
	return bytes.HasPrefix(tx, []byte("zaop"))
}
