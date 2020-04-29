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

import "gopkg.in/urfave/cli.v1"

type AnntoolFlags struct {
	abif,
	callf,
	addr,
	payload,
	bytecode,
	nonce,
	abistr,
	callstr,
	value,
	amount,
	hash,
	accountPubkey,
	peerPubkey,
	validatorPubkey,
	validatorSignature,
	validatorPrivKey,
	power,
	rpc,
	to,
	cType,
	verbose,
	nPrivs,
	pageNum ,
	pageSize ,
	key ,
	codeHash cli.Flag
}

var anntoolFlags = AnntoolFlags{
	cType: cli.StringFlag{
		Name:  "crypto_type",
		Usage: "choose one of the three crypto_type types: \n\t'ZA' includes ed25519,ecdsa,ripemd160,keccak256,secretbox;",
	},
	abif: cli.StringFlag{
		Name:  "abif",
		Usage: "abi definition file",
	},
	callf: cli.StringFlag{
		Name:  "callf",
		Usage: "params file defined in JSON",
	},
	addr: cli.StringFlag{
		Name: "address",
	},
	payload: cli.StringFlag{
		Name: "payload",
	},
	bytecode: cli.StringFlag{
		Name: "bytecode",
	},
	nonce: cli.Uint64Flag{
		Name: "nonce",
	},
	abistr: cli.StringFlag{
		Name: "abi",
	},
	callstr: cli.StringFlag{
		Name: "calljson",
	},
	to: cli.StringFlag{
		Name: "to",
	},
	value: cli.Int64Flag{
		Name: "value",
	},
	amount: cli.Uint64Flag{
		Name: "amount",
	},
	hash: cli.StringFlag{
		Name: "hash",
	},
	accountPubkey: cli.StringFlag{
		Name: "account_pubkey",
	},
	peerPubkey: cli.StringFlag{
		Name: "peer_pubkey",
	},
	validatorPubkey: cli.StringFlag{
		Name: "validator_pubkey",
	},
	validatorSignature: cli.StringFlag{
		Name: "validator_signature",
	},
	validatorPrivKey: cli.StringFlag{
		Name: "validator_privkey",
	},
	power: cli.IntFlag{
		Name: "power",
	},
	rpc: cli.StringFlag{
		Name:  "rpc",
		Value: "tcp://0.0.0.0:46657",
	},
	codeHash: cli.StringFlag{
		Name:  "code_hash",
		Value: "",
	},
	verbose: cli.BoolFlag{
		Name: "verbose",
	},
	nPrivs: cli.IntFlag{
		Name:  "nPrivs",
		Usage: "number of ca privateKey!",
	},
	pageNum:cli.UintFlag{
		Name:"page_num",
	},
	pageSize:cli.UintFlag{
		Name:"page_size",
	},
	key:cli.StringFlag{
		Name:"key",
	},
}
