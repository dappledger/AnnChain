package plugin

import (
	"bytes"
	"errors"
	"fmt"

	"encoding/json"
	anntypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
	//"github.com/dappledger/AnnChain/module/lib/go-wire"
	"go.uber.org/zap"
)

type VoteChannel struct {
	sw          *p2p.Switch
	validators  **anntypes.ValidatorSet
	privkey     crypto.PrivKeyEd25519
	db          db.DB
	eventSwitch anntypes.EventSwitch
	plugins     []IPlugin
	logger      *zap.Logger
}

const (
	uncheckMajor23 = -1
)

func NewVoteChannelPlugin(plugins []IPlugin) *VoteChannel {
	vt := &VoteChannel{}
	// vt.db = db.NewDB("votechannel", config.GetString("db_backend"), config.GetString("db_dir"))
	vt.plugins = plugins
	return vt
}

func (v *VoteChannel) Init(p *InitParams) {
	v.sw = p.Switch
	v.logger = p.Logger
	v.validators = p.Validators
	v.privkey = p.PrivKey
	v.db = p.DB
}

func (v *VoteChannel) SetPlugins(ps []IPlugin) {
	v.plugins = ps
}

func (v *VoteChannel) Reload(p *ReloadParams) {
	v.sw = p.Switch
	v.logger = p.Logger
	v.validators = p.Validators
	v.privkey = p.PrivKey
	v.db = p.DB
}

func (v *VoteChannel) Stop() {
	v.db.Close()
}

func (v *VoteChannel) DeliverTx(tx []byte, i int) (bool, error) {
	if !anntypes.IsVoteChannel(tx) {
		return true, nil
	}
	var cmd anntypes.VoteChannelCmd
	err := json.Unmarshal(anntypes.VoteChannelGetBody(tx), &cmd)
	if err != nil || cmd.CmdCode != anntypes.VoteChannel {
		return true, err
	}
	//process vote request, include: new reqeust, sign ack, exec
	switch cmd.SubCmd {
	case anntypes.VoteChannel_NewRequest:
		err = v.processNewRequest(cmd, anntypes.VoteChannelGetBody(tx))
	case anntypes.VoteChannel_Sign:
		err = v.processSignMessage(cmd)
	case anntypes.VoteChannel_Exec:
		err = v.execVoteResult(cmd)
	}
	if err != nil {
		v.logger.Error("deliver tx error :", zap.Error(err))
	}
	return false, err
}

func (v *VoteChannel) ExecBlock(p *ExecBlockParams) (*ExecBlockReturns, error) {
	for i, tx := range p.Block.Data.ExTxs {
		if _, err := v.DeliverTx(tx, i); err != nil {
			v.logger.Error("plugin votechannel execBlock error", zap.Error(err))
		}
	}
	return nil, nil
}

func (v *VoteChannel) CheckTx(tx []byte) (bool, error) {
	if !anntypes.IsVoteChannel(tx) {
		return true, nil
	}
	var cmd anntypes.VoteChannelCmd
	err := json.Unmarshal(anntypes.VoteChannelGetBody(tx), &cmd)
	if err != nil || cmd.CmdCode != anntypes.VoteChannel {
		return true, err
	}
	return false, nil
}

func (v *VoteChannel) BeginBlock(p *BeginBlockParams) (*BeginBlockReturns, error) {
	return nil, nil
}

func (v *VoteChannel) EndBlock(p *EndBlockParams) (*EndBlockReturns, error) {
	//do nothing ,because all plugin's end block will do in execEndBlockOnPlugins function
	return nil, nil
}

func (v *VoteChannel) Reset() {
}

func (v *VoteChannel) processNewRequest(cmd anntypes.VoteChannelCmd, tx []byte) error {
	// //store to leveldb, just validator can store vote data
	// if !v.checkValidator() {
	// 	log.Debug("node pub key:", v.privkey.PubKey().KeyString())
	// 	return errors.New("node not is validator, cannot do new vote request")
	// }

	//check sender is validator
	// if !(*v.validators).HasAddress(cmd.Sender)
	//save to level db
	v.db.SetSync(cmd.Id, tx)
	return nil
}

func (v *VoteChannel) processSignMessage(cmd anntypes.VoteChannelCmd) error {
	// //only validator can do sign action
	// if !v.checkValidator() {
	// 	log.Debug("node pub key:", v.privkey.PubKey().KeyString())
	// 	return errors.New("node not is validator, cannot do new vote request")
	// }

	if cmd.SubCmd != anntypes.VoteChannel_Sign {
		return errors.New("do sign action error, request must be sign type action")
	}
	//merge sign
	storedvalue := v.db.Get(cmd.Id)
	if storedvalue == nil {
		return errors.New("cannot do sign action, bad key")
	}

	var storedcmd anntypes.VoteChannelCmd
	err := json.Unmarshal(storedvalue, &storedcmd)
	if err != nil {
		return errors.New("cannot read binary bytes when process sign message")
	}
	var sigs [][]byte
	for _, sig := range cmd.Signs {
		for idx, s := range storedcmd.Signs {
			if bytes.Equal(sig, s) {
				break
			}
			if idx == (len(storedcmd.Signs) - 1) {
				sigs = append(sigs, sig)
			}
		}
	}
	storedcmd.Signs = append(storedcmd.Signs, sigs...)

	//store to leveldb
	setvalue, _ := json.Marshal(storedcmd)
	v.db.SetSync(cmd.Id, setvalue)
	return nil
}

func (v *VoteChannel) execVoteResult(cmd anntypes.VoteChannelCmd) error {
	//get cmd from db
	storedvalue := v.db.Get(cmd.Id)
	if storedvalue == nil {
		return errors.New("cannot do exec action, bad key")
	}
	var storedcmd anntypes.VoteChannelCmd
	err := json.Unmarshal(storedvalue, &storedcmd)
	if err != nil {
		return errors.New("cannot read binary bytes when process sign message")
	}
	//check 2/3 vote
	err = v.checkmajor23(storedcmd)
	if err != nil {
		return err
	}
	//先調用相应plugin的deliver函数，再在endblock里执行相应plugin的endblock
	realTx := storedcmd.Msg

	for _, p := range v.plugins {
		p.DeliverTx(realTx, uncheckMajor23)
	}
	//delete this vote channel request
	v.db.Delete(cmd.Id)
	return nil
}

func (v *VoteChannel) checkmajor23(cmd anntypes.VoteChannelCmd) error {
	var major23 int64
	for _, sig := range cmd.Signs {
		sigPubkey := crypto.PubKeyEd25519{}
		copy(sigPubkey[:], sig[:32])
		if (*v.validators).HasAddress(sigPubkey.Address()) {
			_, validator := (*v.validators).GetByAddress(sigPubkey.Address())
			pubKey32 := [32]byte(sigPubkey)
			sig64 := [64]byte{}
			copy(sig64[:], sig[32:])
			if ed25519.Verify(&pubKey32, cmd.Txmsg, &sig64) {
				major23 += validator.VotingPower
			} else {
				v.logger.Info("check major 2/3", zap.String("vote nil", fmt.Sprintf("%X", pubKey32)))
			}
		}
	}
	if major23 <= (*v.validators).TotalVotingPower()*2/3 {
		return errors.New("cannot do vote request, voting power insufficient")
	}
	return nil
}

func (v *VoteChannel) checkValidator() bool {
	if !(*v.validators).HasAddress(v.privkey.PubKey().Address()) {
		return false
	}
	return true
}

// func (v *VoteChannel) GetVoteDB() db.DB {
// return v.db
// }
func (v *VoteChannel) SetEventSwitch(sw anntypes.EventSwitch) {
	v.eventSwitch = sw
}
