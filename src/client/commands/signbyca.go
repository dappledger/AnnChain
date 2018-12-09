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
	"encoding/hex"
	"fmt"
	"strings"

	"gopkg.in/urfave/cli.v1"

	agtypes "github.com/dappledger/AnnChain/angine/types"
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
		},
	}
)

func signCA(ctx *cli.Context) error {
	if !ctx.IsSet("sec") || !ctx.IsSet("pub") {
		return cli.NewExitError("sec or pub is null, exit...", 127)
	}
	sec, pub := ctx.String("sec"), ctx.String("pub")

	secbytes, _ := hex.DecodeString(sec)

	for _, whatever := range strings.Split(pub, ",") {
		whatever = strings.Trim(whatever, " ")
		pubkeyStr := whatever[:64]
		chainID := []byte(whatever[64:])
		pubBytes, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		signature, err := agtypes.SignCA(secbytes, append(pubBytes, chainID...))
		if err != nil {
			return cli.NewExitError(err.Error(), 127)
		}
		fmt.Printf("%s : %s\n", whatever, signature)
	}
	return nil
}

func stringto32byte(key string) (byte32 [32]byte) {
	seckey, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	slice := seckey[:]
	for i := range slice {
		byte32[i] = slice[i]
	}
	return
}
func stringto64byte(key string) (byte64 [64]byte) {
	seckey, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	slice := seckey[:]
	for i := range slice {
		byte64[i] = slice[i]
	}
	return
}
func stringToAnybyte(key string) (bytes []byte) {
	seckey, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	copy(bytes, seckey)
	return
}
func byte64Tobyte(bytes64 [64]byte) (bytes []byte) {
	bytes = make([]byte, 64)
	for i := range bytes64 {
		bytes[i] = bytes64[i]
	}
	return
}
func byte32Tobyte(bytes32 [32]byte) (bytes []byte) {
	bytes = make([]byte, 32)
	for i := range bytes32 {
		bytes[i] = bytes32[i]
	}
	return
}
