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

package state

import (
	"bytes"
	"fmt"
	"path"
	"testing"

	"github.com/tendermint/tendermint/config/tendermint_test"
	crypto "github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib/def"
	db "github.com/dappledger/AnnChain/annlibs/lib/go-db"
	//	. "github.com/tendermint/go-common"
	"github.com/tendermint/abci/example/dummy"
	cfg "github.com/tendermint/go-config"
	"github.com/tendermint/tendermint/proxy"
	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
)

var (
	privKey      = crypto.GenPrivKeyEd25519FromSecret([]byte("handshake_test"))
	chainID      = "handshake_chain"
	nBlocks      = 5
	mempool      = MockMempool{}
	testPartSize = 65536
)

//---------------------------------------
// Test block execution

func TestExecBlock(t *testing.T) {
	// TODO
}

//---------------------------------------
// Test handshake/replay

// Sync from scratch
func TestHandshakeReplayAll(t *testing.T) {
	testHandshakeReplay(t, 0)
}

// Sync many, not from scratch
func TestHandshakeReplaySome(t *testing.T) {
	testHandshakeReplay(t, 1)
}

// Sync from lagging by one
func TestHandshakeReplayOne(t *testing.T) {
	testHandshakeReplay(t, nBlocks-1)
}

// Sync from caught up
func TestHandshakeReplayNone(t *testing.T) {
	testHandshakeReplay(t, nBlocks)
}

// Make some blocks. Start a fresh app and apply n blocks. Then restart the app and sync it up with the remaining blocks
func testHandshakeReplay(t *testing.T, n int) {
	config := tendermint_test.ResetConfig("proxy_test_")

	state, store := stateAndStore(config)
	clientCreator := proxy.NewLocalClientCreator(dummy.NewPersistentDummyApplication(path.Join(config.GetString("db_dir"), "1")))
	clientCreator2 := proxy.NewLocalClientCreator(dummy.NewPersistentDummyApplication(path.Join(config.GetString("db_dir"), "2")))
	proxyApp := proxy.NewAppConns(config, clientCreator, NewHandshaker(config, state, store))
	if _, err := proxyApp.Start(); err != nil {
		t.Fatalf("Error starting proxy app connections: %v", err)
	}
	chain := makeBlockchain(t, proxyApp, state)
	store.chain = chain //
	latestAppHash := state.AppHash
	proxyApp.Stop()

	if n > 0 {
		// start a new app without handshake, play n blocks
		proxyApp = proxy.NewAppConns(config, clientCreator2, nil)
		if _, err := proxyApp.Start(); err != nil {
			t.Fatalf("Error starting proxy app connections: %v", err)
		}
		state2, _ := stateAndStore(config)
		for i := 0; i < n; i++ {
			block := chain[i]
			err := state2.ApplyBlock(nil, proxyApp.Consensus(), block, block.MakePartSet(testPartSize).Header(), mempool)
			if err != nil {
				t.Fatal(err)
			}
		}
		proxyApp.Stop()
	}

	// now start it with the handshake
	handshaker := NewHandshaker(config, state, store)
	proxyApp = proxy.NewAppConns(config, clientCreator2, handshaker)
	if _, err := proxyApp.Start(); err != nil {
		t.Fatalf("Error starting proxy app connections: %v", err)
	}

	// get the latest app hash from the app
	res, err := proxyApp.Query().InfoSync()
	if err != nil {
		t.Fatal(err)
	}

	// the app hash should be synced up
	if !bytes.Equal(latestAppHash, res.LastBlockAppHash) {
		t.Fatalf("Expected app hashes to match after handshake/replay. got %X, expected %X", res.LastBlockAppHash, latestAppHash)
	}

	if handshaker.nBlocks != nBlocks-n {
		t.Fatalf("Expected handshake to sync %d blocks, got %d", nBlocks-n, handshaker.nBlocks)
	}

}

//--------------------------
// utils for making blocks

// make some bogus txs
func txsFunc(blockNum int) (txs []agtypes.Tx) {
	for i := 0; i < 10; i++ {
		txs = append(txs, agtypes.Tx([]byte{byte(blockNum), byte(i)}))
	}
	return txs
}

// sign a commit vote
func signCommit(height, round def.INT, hash []byte, header *pbtypes.PartSetHeader) *pbtypes.Vote {
	vote := &pbtypes.Vote{
		Data: &pbtypes.Vote{
			ValidatorAddress: privKey.PubKey().Address(),
			ValidatorIndex:   0,
			Height:           height,
			Round:            round,
			Type:             pbtypes.VoteType_Precommit,
			BlockID:          &pbtypes.BlockID{hash, header},
		},
	}

	sig := privKey.Sign(agtypes.SignBytes(chainID, vote))
	vote.Signature = sig
	return vote
}

// make a blockchain with one validator
func makeBlockchain(t *testing.T, proxyApp proxy.AppConns, state *State) (blockchain []*agtypes.Block) {

	prevHash := state.LastBlockID.Hash
	lastCommit := new(agtypes.Commit)
	prevParts := agtypes.PartSetHeader{}
	valHash := state.Validators.Hash()
	prevBlockID := agtypes.BlockID{prevHash, prevParts}

	for i := 1; i < nBlocks+1; i++ {
		block, parts := agtypes.MakeBlock(nil, nil, i, chainID, txsFunc(i), lastCommit,
			prevBlockID, valHash, state.AppHash, testPartSize, 0)
		fmt.Println(i)
		fmt.Println(prevBlockID)
		fmt.Println(block.LastBlockID)
		err := state.ApplyBlock(nil, block, block.MakePartSet(testPartSize).Header(), mempool, 1)
		if err != nil {
			t.Fatal(i, err)
		}

		voteSet := agtypes.NewVoteSet(chainID, i, 0, agtypes.VoteTypePrecommit, state.Validators)
		vote := signCommit(i, 0, block.Hash(), parts.Header())
		_, err = voteSet.AddVote(vote)
		if err != nil {
			t.Fatal(err)
		}

		blockchain = append(blockchain, block)
		prevHash = block.Hash()
		prevParts = parts.Header()
		lastCommit = voteSet.MakeCommit()
		prevBlockID = agtypes.BlockID{prevHash, prevParts}
	}
	return blockchain
}

// fresh state and mock store
func stateAndStore(config cfg.Config) (*State, *mockBlockStore) {
	stateDB := db.NewMemDB()
	return MakeGenesisState(stateDB, &agtypes.GenesisDoc{
		ChainID: chainID,
		Validators: []agtypes.GenesisValidator{
			agtypes.GenesisValidator{privKey.PubKey(), 10000, "test"},
		},
		AppHash: nil,
	}), NewMockBlockStore(config, nil)
}

//----------------------------------
// mock block store

type mockBlockStore struct {
	config cfg.Config
	chain  []*agtypes.Block
}

func NewMockBlockStore(config cfg.Config, chain []*agtypes.Block) *mockBlockStore {
	return &mockBlockStore{config, chain}
}

func (bs *mockBlockStore) Height() int                         { return len(bs.chain) }
func (bs *mockBlockStore) LoadBlock(height int) *agtypes.Block { return bs.chain[height-1] }
func (bs *mockBlockStore) LoadBlockMeta(height int) *agtypes.BlockMeta {
	block := bs.chain[height-1]
	return &agtypes.BlockMeta{
		Hash:        block.Hash(),
		Header:      block.Header,
		PartsHeader: block.MakePartSet(bs.config.GetInt("block_part_size")).Header(),
	}
}
