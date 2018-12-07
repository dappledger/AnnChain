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

package p2p

import (
	"math/rand"
	"testing"

	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
)

// Returns an empty dummy peer
func randPeer() *Peer {
	return &Peer{
		Key: RandStr(12),
		NodeInfo: &NodeInfo{
			RemoteAddr: Fmt("%v.%v.%v.%v:46656", rand.Int()%256, rand.Int()%256, rand.Int()%256, rand.Int()%256),
			ListenAddr: Fmt("%v.%v.%v.%v:46656", rand.Int()%256, rand.Int()%256, rand.Int()%256, rand.Int()%256),
		},
	}
}

func TestAddRemoveOne(t *testing.T) {
	peerSet := NewPeerSet()

	peer := randPeer()
	err := peerSet.Add(peer)
	if err != nil {
		t.Errorf("Failed to add new peer")
	}
	if peerSet.Size() != 1 {
		t.Errorf("Failed to add new peer and increment size")
	}

	peerSet.Remove(peer)
	if peerSet.Has(peer.Key) {
		t.Errorf("Failed to remove peer")
	}
	if peerSet.Size() != 0 {
		t.Errorf("Failed to remove peer and decrement size")
	}
}

func TestAddRemoveMany(t *testing.T) {
	peerSet := NewPeerSet()

	peers := []*Peer{}
	N := 100
	for i := 0; i < N; i++ {
		peer := randPeer()
		if err := peerSet.Add(peer); err != nil {
			t.Errorf("Failed to add new peer")
		}
		if peerSet.Size() != i+1 {
			t.Errorf("Failed to add new peer and increment size")
		}
		peers = append(peers, peer)
	}

	for i, peer := range peers {
		peerSet.Remove(peer)
		if peerSet.Has(peer.Key) {
			t.Errorf("Failed to remove peer")
		}
		if peerSet.Size() != len(peers)-i-1 {
			t.Errorf("Failed to remove peer and decrement size")
		}
	}
}
