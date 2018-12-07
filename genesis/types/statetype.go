package types

import (
	"math/big"

	"github.com/dappledger/AnnChain/ann-module/xlib"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
)

// // StateDB is an EVM database for full state querying.
// type AccountDB interface {
// 	GetAccount(ethcmn.Address) Account
// 	CreateAccount(ethcmn.Address) Account

// 	SubBalance(ethcmn.Address, *big.Int, string)
// 	AddBalance(ethcmn.Address, *big.Int, string)
// 	GetBalance(ethcmn.Address) *big.Int

// 	GetNonce(ethcmn.Address) uint64
// 	SetNonce(ethcmn.Address, uint64)

// 	AddRefund(*big.Int)
// 	GetRefund() *big.Int

// 	GetState(ethcmn.Address, ethcmn.Hash) ethcmn.Hash
// 	SetState(ethcmn.Address, ethcmn.Hash, ethcmn.Hash)

// 	Suicide(ethcmn.Address) bool
// 	HasSuicided(ethcmn.Address) bool

// 	// Exist reports whether the given account exists in state.
// 	// Notably this should also return true for suicided accounts.
// 	Exist(ethcmn.Address) bool
// 	// Empty returns whether the given account is empty. Empty
// 	// is defined according to EIP161 (balance = nonce = code = 0).
// 	Empty(ethcmn.Address) bool

// 	RevertToSnapshot(int)
// 	Snapshot() int

// 	// AddLog(*types.Log)
// 	AddPreimage(ethcmn.Hash, []byte)
// }

// Account represents a contract or basic ethereum account.
type Account interface {
	SubBalance(amount *big.Int, log string)
	AddBalance(amount *big.Int, log string)
	SetBalance(*big.Int, string)
	SetNonce(uint64)
	Balance() *big.Int
	Address() ethcmn.Address
	ReturnGas(*big.Int)
	ForEachStorage(cb func(key, value ethcmn.Hash) bool)
	Value() *big.Int
	SetCode(ethcmn.Hash, []byte)
	// FillShow(*ShowAccount)
}

type InflationVotes struct {
	Votes *big.Int
	Dest  ethcmn.Address
}

func (i *InflationVotes) Key() xlib.SortKey {
	return &i.Dest
}

func (i *InflationVotes) Less(data xlib.Sortable) bool {
	if iv, ok := data.(*InflationVotes); ok {
		return i.Votes.Cmp(iv.Votes) > 0
	}
	return false
}
