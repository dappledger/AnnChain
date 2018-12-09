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


package types

import (
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
)

type ICivilTx interface {
	GetPubKey() []byte
	SetPubKey([]byte)
	PopSignature() []byte
	SetSignature([]byte)
	Sender() []byte
}

type CivilTx struct {
	sender []byte

	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`
}

func (t *CivilTx) GetPubKey() []byte {
	return t.PubKey
}

func (t *CivilTx) SetPubKey(pk []byte) {
	t.PubKey = pk
}

func (t *CivilTx) PopSignature() []byte {
	s := t.Signature
	t.Signature = nil
	return s
}

func (t *CivilTx) SetSignature(s []byte) {
	t.Signature = s
}

func (t *CivilTx) Sender() []byte {
	if t.sender != nil || len(t.sender) > 0 {
		return t.sender
	}

	pubkey := crypto.PubKeyEd25519{}
	copy(pubkey[:], t.PubKey)
	t.sender = pubkey.Address()

	return t.sender
}
