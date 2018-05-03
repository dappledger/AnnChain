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
