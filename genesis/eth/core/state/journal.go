// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"math/big"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

type journalEntry interface {
	undo(*StateDB)
}

type journal []journalEntry

type (
	// Changes to the account trie.
	createObjectChange struct {
		account *ethcmn.Address
	}
	resetObjectChange struct {
		prev *StateObject
	}
	suicideChange struct {
		account     *ethcmn.Address
		prev        bool // whether account had already suicided
		prevbalance *big.Int
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *ethcmn.Address
		prev    *big.Int
	}
	entryChange struct {
		account *ethcmn.Address
		prev    uint
	}
	flagChange struct {
		account *ethcmn.Address
		prev    uint8
	}
	thresholdChange struct {
		account *ethcmn.Address
		idex    uint
		prev    uint8
	}
	weightChange struct {
		account *ethcmn.Address
		prev    uint8
	}
	inflationChange struct {
		account *ethcmn.Address
		prev    ethcmn.Address
	}
	nonceChange struct {
		account *ethcmn.Address
		prev    uint64
	}
	storageChange struct {
		account       *ethcmn.Address
		key, prevalue ethcmn.Hash
	}
	codeChange struct {
		account            *ethcmn.Address
		prevcode, prevhash []byte
	}
	// Changes to other state values.
	refundChange struct {
		prev *big.Int
	}
	addLogChange struct {
		txhash ethcmn.Hash
	}
	addPreimageChange struct {
		hash ethcmn.Hash
	}
	touchChange struct {
		account *ethcmn.Address
		prev    bool
	}
)

func (ch codeChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setCode(ethcmn.BytesToHash(ch.prevhash), ch.prevcode)
}
func (ch createObjectChange) undo(s *StateDB) {
	delete(s.stateObjects, *ch.account)
	delete(s.stateObjectsDirty, *ch.account)
}

func (ch resetObjectChange) undo(s *StateDB) {
	s.setStateObject(ch.prev)
}

func (ch suicideChange) undo(s *StateDB) {
	obj := s.GetStateObject(*ch.account)
	if obj != nil {
		obj.suicided = ch.prev
		obj.setBalance(ch.prevbalance, "undo")
	}
}

var ripemd = ethcmn.HexToAddress("0000000000000000000000000000000000000003")

func (ch touchChange) undo(s *StateDB) {
	if !ch.prev && *ch.account != ripemd {
		delete(s.stateObjects, *ch.account)
		delete(s.stateObjectsDirty, *ch.account)
	}
}

func (ch balanceChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setBalance(ch.prev, "undo")
}

func (ch nonceChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setNonce(ch.prev)
}

func (ch storageChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setState(ch.key, ch.prevalue)
}

func (ch refundChange) undo(s *StateDB) {
	s.refund = ch.prev
}

func (ch addLogChange) undo(s *StateDB) {
	logs := s.logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.txhash)
	} else {
		s.logs[ch.txhash] = logs[:len(logs)-1]
	}
}

func (ch addPreimageChange) undo(s *StateDB) {
	delete(s.preimages, ch.hash)
}
