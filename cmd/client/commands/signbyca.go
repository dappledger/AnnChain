package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"gopkg.in/urfave/cli.v1"
)

var (
	SignCommand = cli.Command{
		Name:   "sign",
		Action: signCA,
		Usage:  "nodes who want to join a chain must first get a CA signature from one the CA nodes in that chain",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "sec",
				Usage: "CA node's private key",
				Value: "0",
			},
			cli.StringFlag{
				Name:  "pub",
				Value: "0",
				Usage: "pubkeys of the nodes want to join, comma separated",
			},
			cli.StringFlag{
				Name:  "crypto_type",
				Value: "ZA",
				Usage: "choose one of the three crypto_type types: \n\t'ZA' includes ed25519,ecdsa,ripemd160,keccak256,secretbox;\n",
			},
		},
	}
)

func signCA(ctx *cli.Context) error {
	if !ctx.IsSet("pub") {
		return cli.NewExitError("pub is null, exit...", 127)
	}
	pub := ctx.String("pub")

	privKey, err := requireNodePrivKey(ctx)
	if err != nil {
		return err
	}
	plen := crypto.NodePubkeyLen() * 2 //hex字符串，是bytes的2倍;
	for _, whatever := range strings.Split(pub, ",") {
		whatever = strings.Trim(whatever, " ")
		pubkeyStr := whatever
		var chainID []byte
		if len(whatever) > plen {
			pubkeyStr = whatever[:plen]
			chainID = []byte(whatever[plen:])
		}
		pubBytes, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		signature := privKey.Sign(append(pubBytes, chainID...))
		ss := hex.EncodeToString(crypto.GetNodeSigBytes(signature))
		fmt.Printf("%s : %s\n", whatever, ss)
	}

	return nil
}
