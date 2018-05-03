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


package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tmlibs/db"

	"github.com/dappledger/AnnChain/module/lib/go-crypto"
)

type OrgState struct {
	mtx      sync.Mutex
	database db.DB
	rootHash []byte
	trie     *iavl.IAVLTree

	// dirty just holds everything that will be write down to disk on commit
	dirty map[string]struct{}
	// order
	order []string
	// accountCache is just an effecient improvement
	accountCache map[string]*OrgAccount
}

var (
	ErrAccountExisted      = fmt.Errorf("account already existed")
	ErrAccountNotExist     = fmt.Errorf("account not existed")
	ErrInvalidNonce        = fmt.Errorf("invalid nonce")
	ErrInsufficientBalance = fmt.Errorf("balance is lower than the amount to be wired")
)

func NewOrgState(database db.DB) *OrgState {
	return &OrgState{
		database:     database,
		trie:         iavl.NewIAVLTree(1024, database),
		dirty:        make(map[string]struct{}),
		accountCache: make(map[string]*OrgAccount),
		order:        make([]string, 0),
	}
}

func (os *OrgState) Lock() {
	os.mtx.Lock()
}

func (os *OrgState) Unlock() {
	os.mtx.Unlock()
}

func (os *OrgState) CreateAccount(chainID string, balance int64) (accnt *OrgAccount, err error) {
	os.mtx.Lock()
	// very coarse design
	// here we generate a private key for the chain just by its public info "chainID",
	// so basicly, every one can know this private key, which means it is not so "private".
	priv := crypto.GenPrivKeyEd25519FromSecret([]byte(chainID))
	pubkey := priv.PubKey().(*crypto.PubKeyEd25519)
	if _, ok := os.accountCache[chainID]; ok || os.trie.Has([]byte(chainID)) {
		os.mtx.Unlock()
		return nil, ErrAccountExisted
	}
	accnt = NewOrgAccount(os, chainID, pubkey[:], balance)
	os.accountCache[chainID] = accnt
	os.dirty[chainID] = struct{}{}
	os.order = append(os.order, chainID)
	os.mtx.Unlock()
	return
}

func (os *OrgState) GetAccount(chainID string) (accnt *OrgAccount, err error) {
	var ok bool
	os.mtx.Lock()
	if accnt, ok = os.accountCache[chainID]; ok {
		os.mtx.Unlock()
		return
	}
	_, bytes, exist := os.trie.Get([]byte(chainID))
	if !exist {
		err = ErrAccountNotExist
		os.mtx.Unlock()
		return
	}
	accnt = &OrgAccount{}
	if err := accnt.FromBytes(bytes, os); err != nil {
		os.mtx.Unlock()
		return nil, err
	}

	os.accountCache[chainID] = accnt
	os.mtx.Unlock()
	return
}

func (os *OrgState) ExistAccount(chainID string) bool {
	os.mtx.Lock()
	defer os.mtx.Unlock()

	if _, ok := os.accountCache[chainID]; ok {
		return ok
	}
	return os.trie.Has([]byte(chainID))
}

// RemoveAccount acts in a sync-block way
// remove related bufferred data and remove the account from db immediately
func (os *OrgState) RemoveAccount(chainID string) bool {
	os.mtx.Lock()
	delete(os.accountCache, chainID)
	delete(os.dirty, chainID)
	_, removed := os.trie.Remove([]byte(chainID))
	os.mtx.Unlock()
	return removed
}

// Commit returns the new root bytes
func (os *OrgState) Commit() ([]byte, error) {
	os.mtx.Lock()
	for _, id := range os.order {
		if _, ok := os.accountCache[id]; !ok {
			continue
		}
		os.trie.Set([]byte(os.accountCache[id].ChainID), os.accountCache[id].ToBytes())
	}
	os.rootHash = os.trie.Save()
	os.mtx.Unlock()
	return os.rootHash, nil
}

// Load dumps all the buffer, start every thing from a clean state
func (os *OrgState) Load(root []byte) {
	os.mtx.Lock()
	os.accountCache = make(map[string]*OrgAccount)
	os.dirty = make(map[string]struct{})
	os.order = make([]string, 0)
	os.trie.Load(root)
	os.mtx.Unlock()
}

// Reload works the same as Load, just for semantic purpose
func (os *OrgState) Reload(root []byte) {
	os.Lock()
	os.accountCache = make(map[string]*OrgAccount)
	os.dirty = make(map[string]struct{})
	os.order = make([]string, 0)
	os.trie.Load(root)
	os.Unlock()
}

// CanWire should be called between state.Lock() and state.Unlock()
func (os *OrgState) CanWire(fromAcc *OrgAccount, amount int64, nonce uint64, deliver bool) (bool, error) {
	if fromAcc.Balance < amount {
		return false, ErrInsufficientBalance
	}

	if deliver {
		if fromAcc.Nonce != nonce {
			return false, ErrInvalidNonce
		}
	} else {
		if fromAcc.Nonce > nonce {
			return false, fmt.Errorf("%s,cur:%v,got:%v", ErrInvalidNonce, fromAcc.Nonce, nonce)
		}
	}
	return true, nil
}

func (os *OrgState) Wire(from []byte, to []byte, amount int64, nonce uint64) (int64, error) {
	var (
		fromAcc, toAcc *OrgAccount
		err            error
	)

	if fromAcc, err = os.GetAccount(string(from)); err != nil {
		return 0, err
	}
	if toAcc, err = os.GetAccount(string(to)); err != nil {
		if toAcc, err = os.CreateAccount(string(to), 0); err != nil {
			return 0, err
		}
	}
	os.mtx.Lock()
	if yes, err := os.CanWire(fromAcc, amount, nonce, true); !yes {
		os.mtx.Unlock()
		return 0, err
	}
	_ = toAcc
	// var balance int64
	// if bytes.Equal(from, IssuerAccount.Bytes()) {
	// 	balance = 0
	// } else {
	// 	balance = fromAcc.balance - amount
	// 	fromAcc.nonce++
	// 	os.markDirty(fromAcc)
	// }
	// toAcc.balance.Add(toAcc.balance, amount)
	// s.markDirty(toAcc)
	os.mtx.Unlock()

	return 0, nil
}

// ModifyAccount puts accounts into dirty cache and they will be persisted during commit
func (os *OrgState) ModifyAccount(account *OrgAccount) {
	os.accountCache[account.ChainID] = account
	if _, ok := os.dirty[account.ChainID]; !ok {
		os.dirty[account.ChainID] = struct{}{}
		os.order = append(os.order, account.ChainID)
	}
}

func (os *OrgState) Copy() *OrgState {
	os.mtx.Lock()
	cp := &OrgState{
		database:     os.database,
		rootHash:     os.rootHash,
		trie:         os.trie.Copy().(*iavl.IAVLTree),
		accountCache: make(map[string]*OrgAccount),
		dirty:        make(map[string]struct{}),
		order:        make([]string, len(os.order), cap(os.order)),
	}
	for k := range os.accountCache {
		cp.accountCache[k] = os.accountCache[k].Copy()
	}
	for id := range os.dirty {
		cp.dirty[id] = struct{}{}
	}
	for i, id := range os.order {
		cp.order[i] = id
	}

	os.mtx.Unlock()
	return cp
}

// OrgAccount abstracts account model
// Pay attention that you have to call orgstate.ModifyAccount when you change anything in the account
// otherwise, your changes won't live across commits.
type OrgAccount struct {
	master *OrgState

	ChainID string                       `json:"chainid"`
	PubKey  []byte                       `json:"pubkey"`
	Nonce   uint64                       `json:"nonce"`
	Balance int64                        `json:"balance"`
	Count   int                          `json:"count"`
	Nodes   map[string]map[string]string `json:"nodes"`
}

// NewOrgAccount also binds org state with the newly created account
func NewOrgAccount(state *OrgState, chainID string, pubkey []byte, balance int64) *OrgAccount {
	return &OrgAccount{
		master: state,

		ChainID: chainID,
		Nonce:   0,
		PubKey:  pubkey,
		Balance: balance,
		Count:   0,
		Nodes:   make(map[string]map[string]string),
	}
}

// FromBytes also restores the connection between the account and the org state
func (oa *OrgAccount) FromBytes(bytes []byte, state *OrgState) error {
	if err := json.Unmarshal(bytes, oa); err != nil {
		return err
	}
	oa.master = state
	return nil
}

func (oa *OrgAccount) ToBytes() []byte {
	bys, err := json.Marshal(oa)
	if err != nil {
		return nil
	}
	return bys
}

func (oa *OrgAccount) GetNonce() uint64 {
	return oa.Nonce
}

func (oa *OrgAccount) SetNonce(n uint64) {
	oa.Nonce = n
	oa.master.ModifyAccount(oa)
}

func (oa *OrgAccount) GetBalance() int64 {
	return oa.Balance
}

func (oa *OrgAccount) SetBalance(b int64) {
	oa.Balance = b
	oa.master.ModifyAccount(oa)
}

func (oa *OrgAccount) GetPubkey() []byte {
	return oa.PubKey
}

func (oa *OrgAccount) GetPubkeyString() string {
	return hex.EncodeToString(oa.PubKey)
}

func (oa *OrgAccount) GetChainID() string {
	return oa.ChainID
}

func (oa *OrgAccount) GetNodes() map[string]map[string]string {
	return oa.Nodes
}

func (oa *OrgAccount) AddNode(p crypto.PubKey, attr map[string]string) error {
	pubkey, ok := p.(*crypto.PubKeyEd25519)
	if !ok {
		return fmt.Errorf("only support Ed25519")
	}
	if _, ok := oa.Nodes[pubkey.KeyString()]; ok {
		return fmt.Errorf("pubkey already exists")
	}
	oa.Nodes[pubkey.KeyString()] = attr
	oa.Count++
	oa.master.ModifyAccount(oa)
	return nil
}

func (oa *OrgAccount) RemoveNode(p crypto.PubKey) error {
	pubkey, ok := p.(*crypto.PubKeyEd25519)
	if !ok {
		return fmt.Errorf("only support Ed25519")
	}
	if _, ok := oa.Nodes[pubkey.KeyString()]; !ok {
		return fmt.Errorf("no such node")
	}
	delete(oa.Nodes, pubkey.KeyString())
	oa.Count--
	oa.master.ModifyAccount(oa)
	return nil
}

func (oa *OrgAccount) Copy() *OrgAccount {
	cp := OrgAccount{
		master:  oa.master,
		ChainID: oa.ChainID,
		Nonce:   oa.Nonce,
		PubKey:  oa.PubKey,
		Balance: oa.Balance,
		Count:   oa.Count,
		Nodes:   make(map[string]map[string]string),
	}
	for k, a := range oa.Nodes {
		copyAttrs := make(map[string]string)
		for an, av := range a {
			copyAttrs[an] = av
		}
		cp.Nodes[k] = copyAttrs
	}
	return &cp
}
