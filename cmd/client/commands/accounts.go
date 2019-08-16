// Copyright © 2017 ZhongAn Technology
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
	privkeyBytes, addrBytes, err := createAccount()
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	fmt.Printf("privkey: %X\n", privkeyBytes)
	fmt.Printf("address: %X\n", addrBytes)
	return nil
}

func createAccount() ([]byte, []byte, error) {
	var (
		privkeyBytes []byte
		addrBytes    []byte
	)

	privkey, err := crypto.GenerateKey()
	if err != nil {
		return privkeyBytes, addrBytes, cli.NewExitError(err.Error(), 127)
	}

	privkeyBytes = crypto.FromECDSA(privkey)

	address := crypto.PubkeyToAddress(privkey.PublicKey)
	addrBytes = address.Bytes()

	return privkeyBytes, addrBytes, nil
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
	privkeyStr = key

	fmt.Println("privkey: ", privkeyStr)
	fmt.Println("pubkey:", pubkeyStr)
	fmt.Println("address: ", addrStr)
	return nil
}
