package types

import "bytes"

type VoteChannelCmd struct {
	Id       []byte   `json:"id"`
	CmdCode  string   `json:"cmdcode"`  //check for vote channel op
	SubCmd   string   `json:"subcmd"`   //check for request type, include new vote, exec, vote ...
	Votetype string   `json:"votetype"` //check is SpecialOP or other kind of OP type
	Txmsg    Tx       `json:"txmsg"`
	Msg      []byte   `json:"msg"` //store any kind of specific request type
	Signs    [][]byte `json:"signs"`
	Sender   []byte   `json:"node_pubkey"` //sender pubkey
}

var Votetag = []byte{'v', 'o', 't', 0x01}

//sub command
const (
	VoteChannel            = "votechannel"
	VoteChannel_NewRequest = "newVoteRequest"
	VoteChannel_Sign       = "signForRequest"
	VoteChannel_Exec       = "executeRequest"
	VoteChannel_query      = "queryRequests"
)

func TagVoteChannelTx(tx []byte) []byte {
	return append(Votetag, tx...)
}

func IsVoteChannel(tx []byte) bool {
	return bytes.HasPrefix(tx, Votetag)
}

func VoteChannelGetBody(tx []byte) []byte {
	if len(tx) > 4 {
		return tx[4:]
	}
	return tx
}
