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


package state

import (
	"math/big"

	"go.uber.org/zap"

	"github.com/tendermint/tmlibs/db"

	cvtools "github.com/dappledger/AnnChain/src/tools"
	cvtypes "github.com/dappledger/AnnChain/src/types"
)

type RemoteState struct {
	*State
	revert *RevertLayer
	log    *zap.Logger
}

func NewRemoteState(database db.DB, log *zap.Logger) *RemoteState {
	rs := &RemoteState{
		State:  NewState(database),
		log:    log,
		revert: &RevertLayer{},
	}
	rs.revert.Init(rs)
	return rs
}

func (rs *RemoteState) RemoteAcc(accID string) *RemoteAccount {
	acc, err := rs.State.GetData(accID, RemoteAccFromBytes)
	if err != nil {
		return nil
	}
	ra, ok := acc.(*RemoteAccount)
	if !ok {
		return nil
	}
	ra.Load(rs)
	return ra
}

func (rs *RemoteState) Copy() *RemoteState {
	cprs := &RemoteState{
		State:  rs.State.Copy(),
		revert: rs.revert.Copy(),
		log:    rs.log,
	}
	return cprs
}

func (rs *RemoteState) Snapshot() int {
	return rs.revert.Snapshot()
}

func (rs *RemoteState) RevertToSnapshot(version int) {
	rs.revert.RevertToVerstion(version)
}

func (rs *RemoteState) CreateRemoteAcc(accID string, balance *big.Int) *RemoteAccount {
	var remote RemoteAccount
	remote.Init(accID, balance, rs)
	if err := rs.State.CreateData(&remote); err != nil {
		rs.log.Warn("[remote_state],create remote account err", zap.Error(err))
		return nil
	}
	return &remote
}

func (rs *RemoteState) Transfer(fromID, toID string, balance *big.Int) bool {
	if balance.Cmp(cvtypes.BigInt0()) == 0 {
		return true
	}
	from := rs.RemoteAcc(fromID)
	if from == nil {
		rs.log.Warn("[remote_state],transfer from not found", zap.String("from", fromID), zap.String("to", toID))
		return false
	}
	if err := from.SubBalance(balance); err != nil {
		rs.log.Warn("[remote_state],transfer from got err when sub balance", zap.String("from", fromID), zap.String("to", toID), zap.Error(err))
		return false
	}
	to := rs.RemoteAcc(toID)
	if to == nil {
		to = rs.CreateRemoteAcc(toID, cvtypes.BigInt0())
		rs.log.Info("[remote_state],create account", zap.String("from", fromID), zap.String("create", toID))
	}
	to.AddBalance(balance)
	rs.log.Info("[remote_state],transfer", zap.String("from", fromID), zap.String("to", toID), zap.String("balance", balance.String()), zap.String("from_balance", from.GetBalance().String()), zap.String("to_balance", to.GetBalance().String()))
	return true
}

func (rs *RemoteState) GetNonce(accID string) uint64 {
	acc := rs.RemoteAcc(accID)
	if acc == nil {
		return 0
	}
	return acc.GetNonce()
}

func (rs *RemoteState) AddNonce(accID string, add uint64) {
	acc := rs.RemoteAcc(accID)
	if acc == nil {
		acc = rs.CreateRemoteAcc(accID, cvtypes.BigInt0())
		rs.log.Warn("[remote_state],add nonce,account doesn't exist,create account", zap.String("acc", accID))
	}
	acc.AddNonce(add)
}

func (rs *RemoteState) GetBalance(accID string) *big.Int {
	acc := rs.RemoteAcc(accID)
	if acc == nil {
		return cvtypes.BigInt0()
	}
	return acc.GetBalance()
}

func (rs *RemoteState) AccSetKv(accID string, key string, value []byte) error {
	acc := rs.RemoteAcc(accID)
	if acc == nil {
		acc = rs.CreateRemoteAcc(accID, cvtypes.BigInt0())
		rs.log.Warn("[remote_state],set kv,account doesn't exist,create account", zap.String("acc", accID))
	}
	return acc.SetKv(key, value)
}

func (rs *RemoteState) AccGetKv(accID string, key string) []byte {
	acc := rs.RemoteAcc(accID)
	if acc == nil {
		return nil
	}
	return acc.GetKv(key)
}

func (rs *RemoteState) AccDelKv(accID string, key string) {
	acc := rs.RemoteAcc(accID)
	if acc == nil {
		return
	}
	acc.DelKv(key)
	return
}

//=======================RemoteAccount==============================

type RemoteAccount struct {
	cvtypes.RemoteAccountData

	state   *RemoteState
	balance *big.Int
	store   *State
}

func (ra *RemoteAccount) Init(accID string, balance *big.Int, state *RemoteState) {
	ra.ID = accID
	ra.balance = big.NewInt(0).Set(balance)
	ra.store = NewState(state.database)
	ra.state = state
}

func (ra *RemoteAccount) Load(state *RemoteState) {
	if ra.state == nil {
		ra.state = state
		ra.store = NewState(state.database)
		ra.store.Load(ra.StorageRoot)
	}
}

func (ra *RemoteAccount) Copy() cvtypes.StateDataItfc {
	ret := *ra
	ret.store = ra.store.Copy()
	ret.state = ra.state
	ret.balance = big.NewInt(0).Set(ra.GetBalance())
	return &ret
}

func (ra *RemoteAccount) onModified() {
	ra.state.ModifyData(ra.Key(), ra)
}

func (ra *RemoteAccount) AddNonce(add uint64) {
	ra.state.revert.addJourney(&ModifyNonceUndo{
		accID:    ra.Key(),
		preNonce: ra.Nonce,
	})
	ra.addNonce(add)
}

func (ra *RemoteAccount) addNonce(add uint64) {
	ra.Nonce += add
	ra.onModified()
}

func (ra *RemoteAccount) setNonce(num uint64) {
	ra.Nonce = num
	ra.onModified()
}

func (ra *RemoteAccount) GetNonce() uint64 {
	return ra.Nonce
}

func (ra *RemoteAccount) AddBalance(add *big.Int) {
	ra.state.revert.addJourney(&ModifyBalanceUndo{
		accID:      ra.Key(),
		preBalance: ra.GetBalance(),
	})
	ra.addBalance(add)
}

func (ra *RemoteAccount) addBalance(add *big.Int) {
	curBalance := ra.GetBalance()
	ra.balance = curBalance.Add(curBalance, add)
	ra.onModified()
}

func (ra *RemoteAccount) SubBalance(sub *big.Int) error {
	if ra.GetBalance().Cmp(sub) < 0 {
		return ErrInsufficientBalance
	}
	ra.state.revert.addJourney(&ModifyBalanceUndo{
		accID:      ra.Key(),
		preBalance: ra.GetBalance(),
	})
	ra.subBalance(sub)
	return nil
}

func (ra *RemoteAccount) subBalance(sub *big.Int) {
	curBalance := ra.GetBalance()
	ra.balance = curBalance.Sub(curBalance, sub)
	ra.onModified()
}

func (ra *RemoteAccount) setBalance(blc *big.Int) {
	ra.balance = big.NewInt(0).Set(blc)
	ra.onModified()
}

func (ra *RemoteAccount) GetBalance() *big.Int {
	if ra.balance == nil {
		ra.balance = big.NewInt(0)
		ra.balance.SetBytes(ra.Blcbys)
	}
	return ra.balance
}

func (ra *RemoteAccount) Key() string {
	return ra.ID
}

func (ra *RemoteAccount) Bytes() ([]byte, error) {
	ra.Blcbys = ra.GetBalance().Bytes()
	return cvtools.PbMarshal(&ra.RemoteAccountData), nil
}

func (ra *RemoteAccount) OnCommit() error {
	root, err := ra.store.Commit()
	if err != nil {
		return err
	}
	ra.StorageRoot = root
	return nil
}

func (ra *RemoteAccount) SetKv(key string, value []byte) (err error) {
	rkey := JointKvDataKey(ra.Key(), key)
	preData := ra.getKv(rkey)
	if err = ra.setKv(rkey, value); err == nil {
		ra.state.revert.addJourney(&ModifyKvUndo{
			accID:   ra.Key(),
			key:     rkey,
			preData: preData,
		})
	}
	return
}

func (ra *RemoteAccount) setKv(key string, value []byte) (err error) {
	var kvdata StateKvData
	kvdata.Init(key, value)
	if ra.store.ExistData(key) {
		ra.store.ModifyData(kvdata.Key(), &kvdata)
	} else {
		err = ra.store.CreateData(&kvdata)
	}
	ra.onModified()
	return
}

func (ra *RemoteAccount) DelKv(key string) {
	rkey := JointKvDataKey(ra.Key(), key)
	preData := ra.getKv(rkey)
	if len(preData) > 0 {
		ra.state.revert.addJourney(&ModifyKvUndo{
			accID:   ra.Key(),
			key:     rkey,
			preData: preData,
		})
		ra.delKv(rkey)
	}
}

func (ra *RemoteAccount) delKv(key string) {
	ra.store.ModifyData(key, nil)
	ra.onModified()
}

func (ra *RemoteAccount) GetKv(key string) []byte {
	rkey := JointKvDataKey(ra.Key(), key)
	return ra.getKv(rkey)
}

func (ra *RemoteAccount) getKv(rkey string) []byte {
	res, err := ra.store.GetData(rkey, KVDataFromBytes)
	if err != nil {
		//ra.state.log.Warn("[remote_account],get kv err", zap.Error(err))
		return nil
	}
	data, _ := res.(*StateKvData)
	bytes, _ := data.Bytes()
	return bytes
}

//=======================Undo effectuation==============================

type ModifyNonceUndo struct {
	accID    string
	preNonce uint64
}

func (n *ModifyNonceUndo) Undo(state cvtypes.StateItfc) {
	rstt := state.(*RemoteState)
	acc := rstt.RemoteAcc(n.accID)
	acc.setNonce(n.preNonce)
}

func (n *ModifyNonceUndo) Copy() ModifiedUndoItfc {
	return &ModifyNonceUndo{
		accID:    n.accID,
		preNonce: n.preNonce,
	}
}

type ModifyBalanceUndo struct {
	accID      string
	preBalance *big.Int
}

func (n *ModifyBalanceUndo) Undo(state cvtypes.StateItfc) {
	rstt := state.(*RemoteState)
	acc := rstt.RemoteAcc(n.accID)
	acc.setBalance(n.preBalance)
}

func (n *ModifyBalanceUndo) Copy() ModifiedUndoItfc {
	return &ModifyBalanceUndo{
		accID:      n.accID,
		preBalance: big.NewInt(0).Set(n.preBalance),
	}
}

type ModifyKvUndo struct {
	accID   string
	key     string
	preData []byte
}

func (n *ModifyKvUndo) Undo(state cvtypes.StateItfc) {
	rstt := state.(*RemoteState)
	acc := rstt.RemoteAcc(n.accID)
	acc.setKv(n.key, n.preData)
}

func (n *ModifyKvUndo) Copy() ModifiedUndoItfc {
	return &ModifyKvUndo{
		accID:   n.accID,
		key:     n.key,
		preData: cvtools.CopyBytes(n.preData),
	}
}
