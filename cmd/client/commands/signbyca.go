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
	"encoding/hex"
	"fmt"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
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
				Usage: "choose one of the three crypto_type types: \n\t'ZA' includes ed25519,ecdsa,ripemd160,keccak256,secretbox;",
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
