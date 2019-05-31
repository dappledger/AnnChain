package commands

import (
	"encoding/hex"
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
)

func requirePubKey(ctx *cli.Context) (crypto.PubKey, error) {
	crypto.NodeInit(ctx.String("crypto_type"))
	if !ctx.IsSet("pubkey") {
		return nil, fmt.Errorf("pubkey is required")
	}
	pub := ctx.String("pubkey")
	pubBytes, err := hex.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	pubkey := crypto.SetNodePubkey(pubBytes)
	return pubkey, nil
}

func requireNodePrivKey(ctx *cli.Context) (crypto.PrivKey, error) {
	crypto.NodeInit(ctx.String("crypto_type"))

	var privkey string

	if privkey = ctx.String("sec"); privkey == "" {

		fmt.Println("Privkey for user :")
		bytePriv, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, cli.NewExitError("fail to read privkey", 127)
		}
		privkey = string(bytePriv)
	}

	if privkey == "" {
		return nil, cli.NewExitError("privkey is required", 127)
	}

	if strings.Index(privkey, "0x") == 0 {
		privkey = privkey[2:]
	}
	if strings.Index(privkey, "0X") == 0 {
		privkey = privkey[2:]
	}

	privBytes, err := hex.DecodeString(privkey)
	if err != nil {
		return nil, err
	}
	key := crypto.SetNodePrivKey(privBytes)
	return key, nil
}

func requireAccPrivky(ctx *cli.Context) (string, error) {
	var privkey string

	fmt.Println("Privkey for user :")
	bytePriv, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", cli.NewExitError("fail to read privkey", 127)
	}

	privkey = string(bytePriv)

	if privkey == "" {
		return "", cli.NewExitError("privkey is required", 127)
	}

	if strings.Index(privkey, "0x") == 0 {
		privkey = privkey[2:]
	}

	return privkey, nil
}
