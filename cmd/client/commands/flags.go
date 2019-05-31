package commands

import (
	"gopkg.in/urfave/cli.v1"
)

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
	power,
	isCA,
	rpc,
	to,
	cType,
	codeHash cli.Flag
}

var anntoolFlags = AnntoolFlags{
	cType: cli.StringFlag{
		Name:  "crypto_type",
		Usage: "choose one of the three crypto_type types: \n\t'ZA' includes ed25519,ecdsa,ripemd160,keccak256,secretbox;\n",
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
	power: cli.IntFlag{
		Name: "power",
	},
	isCA: cli.BoolFlag{
		Name: "isCA",
	},
	rpc: cli.StringFlag{
		Name:  "rpc",
		Value: "tcp://0.0.0.0:46657",
	},
	codeHash: cli.StringFlag{
		Name:  "code_hash",
		Value: "",
	},
}
