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

func requireNodePrivKeys(ctx *cli.Context) ([]crypto.PrivKey, error) {
	crypto.NodeInit(ctx.String("crypto_type"))
	nPriv := 1
	if ctx.IsSet(NPrivs()) {
		nPriv = ctx.Int(NPrivs())
	}
	fmt.Printf("need %d private keys;", nPriv)
	var rets []crypto.PrivKey
	for i := 0; i < nPriv; i++ {
		fmt.Printf(" please input %d' keys:\n", i+1)
		key, err := readNodePrivKey(ctx)
		if err != nil {
			return nil, err
		}
		if ctx.IsSet(Verbose()) && ctx.Bool(Verbose()) {
			fmt.Printf("fetch privkey of addr(%x)\n", key.PubKey().Address())
		}
		rets = append(rets, key)
	}

	return rets, nil
}

func requireNodePrivKey(ctx *cli.Context) (crypto.PrivKey, error) {
	crypto.NodeInit(ctx.String("crypto_type"))
	fmt.Println("Node Privkey for user:")
	key, err := readNodePrivKey(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.IsSet(Verbose()) && ctx.Bool(Verbose()) {
		fmt.Printf("fetch privkey of addr(%x)\n", key.PubKey().Address())
	}
	return key, err
}

func readNodePrivKey(ctx *cli.Context) (crypto.PrivKey, error) {
	privkeybytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, cli.NewExitError("fail to read privkey", 127)
	}
	privkey := string(privkeybytes)
	if privkey[0:2] == "0x" || privkey[0:2] == "0X" {
		privkey = privkey[2:]
	}

	privBytes, err := hex.DecodeString(string(privkey))
	if err != nil {
		fmt.Printf("DecodeString(prvkey)=%s;err=%s\n", privkey, err.Error())
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

	if strings.Index(privkey, "0x") == 0 || strings.Index(privkey, "0X") == 0 {
		privkey = privkey[2:]
	}

	return privkey, nil
}
