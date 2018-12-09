package angine

import (
	"bytes"
	"errors"

	"encoding/json"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	atypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
)

//RequestVoteChannel handle user rpc requst
func (e *Angine) ProcessVoteChannel(tx atypes.Tx) (*atypes.ResultRequestVoteChannel, error) {
	//case send new reqeust, sign, get info , exec
	res := &atypes.ResultRequestVoteChannel{
		Code: pbtypes.CodeType_InternalError,
	}
	//check if this node is validator node, only validator can do this kind of operation
	_, validators := e.consensus.GetValidators()
	ourPubkey := e.privValidator.GetPubKey().(*crypto.PubKeyEd25519)
	for idx, val := range validators {
		if val.PubKey.KeyString() == ourPubkey.KeyString() {
			break
		}
		if idx == (len(validators) - 1) {
			res.Log = "only validator can do vote command action"
			return res, errors.New(res.Log)
		}
	}
	//check vote channel
	if !atypes.IsVoteChannel(tx) {
		res.Log = "Not type Vote Channel request"
		return res, errors.New(res.Log)
	}
	var votecmd atypes.VoteChannelCmd
	err := json.Unmarshal(atypes.VoteChannelGetBody(tx), &votecmd)
	if err != nil {
		res.Log = "error when unmarshal tx"
		return res, err
	}

	if votecmd.SubCmd == atypes.VoteChannel_NewRequest {
		//which node recieve this rpc request, so it will be sender node, need to know, only validator node can send new vote reqeust
		votecmd.Sender = ourPubkey.Bytes() //use sender to check whether new vote request is validator node made
		votecmd.Txmsg = tx
		signature := e.privValidator.GetPrivKey().Sign(tx)
		votecmd.Id = signature.Bytes()
		votecmd.Signs = append(votecmd.Signs, append(e.privValidator.GetPubKey().Bytes(), signature.Bytes()...))
	} else if votecmd.SubCmd == atypes.VoteChannel_Sign {
		err = e.handleRequestVoteSign(&votecmd)
	} else if votecmd.SubCmd == atypes.VoteChannel_Exec {
		err = e.handleRequestVoteExec(votecmd)
	} else if votecmd.SubCmd == atypes.VoteChannel_query { //show vote reqeusts
		b, err := e.handleRequestQuery()
		if err == nil {
			return &atypes.ResultRequestVoteChannel{
				Code: pbtypes.CodeType_OK,
				Data: b,
				Log:  "success",
			}, err
		}
	} else {
		res.Log = "unsupported vote sub command"
		return res, errors.New(res.Log)
	}
	if err != nil {
		res.Log = err.Error()
		return res, err
	}

	btx, _ := json.Marshal(votecmd)
	err = e.BroadcastTx(atypes.WrapTx(atypes.Votetag, btx))
	if err != nil {
		return &atypes.ResultRequestVoteChannel{
			Code: pbtypes.CodeType_InternalError,
			Log:  err.Error(),
		}, err
	}
	return &atypes.ResultRequestVoteChannel{
		Code: pbtypes.CodeType_OK,
	}, err
}

func (e *Angine) handleRequestVoteSign(cmd *atypes.VoteChannelCmd) error {
	ourPubkey := e.privValidator.GetPubKey().(*crypto.PubKeyEd25519)
	//get request from leveldb
	value := e.dbs["votechannel"].Get((*cmd).Id)
	if value == nil {
		return errors.New("cannot do this action, becasuse node have not store this request yet")
	}
	//签名时，在signs字段前加上此节点公钥字符, 当节点投票为不赞成的票，那么client 傳來Txmsg为"reject"的byte轉換，一定会导致验证失败，此时验证签名时，不会加上此节点的voting power
	var storedcmd atypes.VoteChannelCmd
	err := json.Unmarshal(value, &storedcmd)
	if err != nil {
		return err
	}
	for _, sig := range storedcmd.Signs {
		pk := crypto.PubKeyEd25519{}
		copy(pk[:], sig[:32])
		if bytes.Equal(pk.Bytes(), ourPubkey.Bytes()) {
			return errors.New("you have made sign action before")
		}
	}

	cmdsig := crypto.SignatureEd25519{}
	copy(cmdsig[:], (*cmd).Txmsg)
	sigBytes := [64]byte(cmdsig)
	valPubKey := [32]byte(*ourPubkey)
	if ed25519.Verify(&valPubKey, (*cmd).Id, &sigBytes) {
		(*cmd).Txmsg = storedcmd.Txmsg
	} else {
		(*cmd).Txmsg = []byte("reject")
	}

	//append node sign
	mysigbytes := append(ourPubkey.Bytes(), e.privValidator.GetPrivKey().Sign((*cmd).Txmsg).Bytes()...)
	(*cmd).Signs = append((*cmd).Signs, mysigbytes)

	return nil
}

func (e *Angine) handleRequestVoteExec(cmd atypes.VoteChannelCmd) error {
	ourPubkey := e.privValidator.GetPubKey().(*crypto.PubKeyEd25519)
	value := e.dbs["votechannel"].Get(cmd.Id)
	if value == nil {
		return errors.New("cannot do exec, becasuse node have not store this request yet")
	}
	var storedcmd atypes.VoteChannelCmd
	json.Unmarshal(value, &storedcmd)

	//check if node is request sender, only origin request sender can do this kind of reqeust
	if !bytes.Equal(ourPubkey.Bytes(), storedcmd.Sender) {
		return errors.New("node not is origin requester")
	}

	//check if 2/3 voting power is ok
	//_, validators := consensusState.GetValidators()
	_, validators := e.consensus.GetValidators()
	totalvp := e.consensus.GetTotalVotingPower()
	var major23 int64
	for _, validator := range validators {
		for _, sig := range storedcmd.Signs {
			sigPubKey := crypto.PubKeyEd25519{}
			copy(sigPubKey[:], sig[:32])
			//sigPubKey, err := crypto.PubKeyFromBytes(sig[:33])
			if validator.PubKey.Equals(&sigPubKey) {
				valPubKey := [32]byte(*validator.PubKey.PubKey.(*crypto.PubKeyEd25519))
				signature := crypto.SignatureEd25519{}
				copy(signature[:], sig[32:])
				sigByte64 := [64]byte(signature)
				if ed25519.Verify(&valPubKey, storedcmd.Txmsg, &sigByte64) {
					major23 += validator.VotingPower
				}
				break
			}
		}
	}
	if major23 <= totalvp*2/3 {
		return errors.New("cannot do exec vote request, voting power insufficient")
	}
	return nil
}

func (e *Angine) handleRequestQuery() ([]byte, error) {
	//get all requsts from db, include id , state, content, origin sender
	queryRes := make([]atypes.VoteChannelCmd, 0)
	iter := e.dbs["votechannel"].Iterator()
	for iter.Next() {
		var cmd atypes.VoteChannelCmd
		tx := iter.Value()
		json.Unmarshal(tx, &cmd)
		queryRes = append(queryRes, cmd)
	}
	b, err := json.Marshal(queryRes)
	if err != nil {
		return nil, err
	}
	return b, nil
}
