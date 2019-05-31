package commands

import (
	"fmt"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/crypto"
)

var (
	//AccountCommands defines a more git-like subcommand system
	AccountCommands = cli.Command{
		Name:     "account",
		Usage:    "operations for account",
		Category: "Account",
		Subcommands: []cli.Command{
			{
				Name:     "create",
				Action:   accountCreate,
				Usage:    "generate new private-pub key pair",
				Category: "Account",
			},
			{
				Name:     "calc",
				Action:   queryPubAddr,
				Usage:    "calculate pubkey and addr from given privkey",
				Category: "Account",
			},
		},
	}
)

func accountCreate(ctx *cli.Context) error {
	var (
		privkeyStr string
		addrStr    string
	)

	privkey, err := crypto.GenerateKey()
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	privkeyStr = common.Bytes2Hex(crypto.FromECDSA(privkey))

	address := crypto.PubkeyToAddress(privkey.PublicKey)
	addrStr = address.Hex()

	fmt.Println("privkey: ", privkeyStr)
	fmt.Println("address: ", addrStr)
	return nil
}

func queryPubAddr(ctx *cli.Context) error {
	var (
		privkeyStr string
		pubkeyStr  string
		addrStr    string
	)

	key, err := requireAccPrivky(ctx)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	priv, err := crypto.HexToECDSA(key)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	privkey := crypto.FromECDSA(priv)
	pubkeyStr = common.Bytes2Hex(privkey)
	pubkeyStr = common.Bytes2Hex(crypto.FromECDSAPub(&priv.PublicKey))

	addrBytes := crypto.PubkeyToAddress(priv.PublicKey)
	addrStr = common.Bytes2Hex(addrBytes[:])

	fmt.Println("privkey: ", privkeyStr)
	fmt.Println("pubkey:", pubkeyStr)
	fmt.Println("address: ", addrStr)
	return nil
}
