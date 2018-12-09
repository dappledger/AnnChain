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
	"fmt"
	"testing"

	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/src/tools"
	civiltypes "github.com/dappledger/AnnChain/src/types"
)

const (
	PrivateKey = "9F04A3EB2E3B412617F0A9D39466B357EBD3A073C28D004C73E482544515898D0FC4E216FB4B40781CEFAECB6C359BA6549069475B7DD678AECF1DF4AC5FCB4E"
)

var (
	priv crypto.PrivKeyEd25519
	pub  crypto.PubKeyEd25519
)

type DummyEventTx struct {
	civiltypes.CivilTx
}

func init() {
	privBytes, _ := hex.DecodeString(PrivateKey)
	copy(priv[:], privBytes)
	pub = priv.PubKey().(crypto.PubKeyEd25519)
}

func TestSign(t *testing.T) {
	event1 := &DummyEventTx{}

	if _, err := tools.TxSign(event1, priv); err != nil {
		panic(err)
	}

	if ok, err := tools.TxVerifySignature(event1); !ok {
		panic(err)
	}

	fmt.Println(event1.Sender())

}
