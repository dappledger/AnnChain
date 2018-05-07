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

package commands

import (
	"fmt"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/crypto"
	"gopkg.in/urfave/cli.v1"
)

var (
	//AccountCommands defines a more git-like subcommand system
	AccountCommands = cli.Command{
		Name:     "account",
		Usage:    "operations for account",
		Category: "Account",
		Subcommands: []cli.Command{
			{
				Name:     "generate",
				Action:   generatePrivPubAddr,
				Usage:    "generate new private-pub key pair",
				Category: "Account",
			},
		},
	}
)

func generatePrivPubAddr(ctx *cli.Context) error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return cli.NewExitError("fail to generate key", 127)
	}

	privkey := crypto.FromECDSA(key)
	pubkey := crypto.FromECDSAPub(&key.PublicKey)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("privkey: ", common.Bytes2Hex(privkey))
	fmt.Println("pubkey:", common.Bytes2Hex(pubkey))
	fmt.Println("address: ", common.Bytes2Hex(addr[:]))
	return nil
}

func addAccount(ctx *cli.Context) error {
	//manager := accounts.NewManager(ctx.GlobalString("dir"), accounts.StandardScryptN, accounts.StandardScryptP)
	// manager := accounts.NewManager()
	// account, err := manager.NewAccount(ac.SanitizeHex(ctx.String("passwd")))
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("Created account at address:", account.Address.Hex())
	return nil
}
