package tools

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/core/state"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	ethcrypto "github.com/dappledger/AnnChain/eth/crypto"
)

var (
	InitPrivateKeyHex       = "0000000000000000000000000000000000000000000000000000000000000000"
	NumAccounts       int64 = 100000
	AccountsPerRound  int64 = 10000
	EthSigner               = ethtypes.HomesteadSigner{}
)

// CoinSetupTestBase just create 1 million accounts with 10000 balance
// without tx, it should be much faster
func CoinSetupTestBase(statedb *state.StateDB) (common.Hash, time.Duration) {
	var root common.Hash
	commitBegin := time.Now()
	for i := 0; i < int(NumAccounts/AccountsPerRound); i++ {
		privkeys := PreparePrivateKeys(i, AccountsPerRound)
		for _, k := range privkeys {
			addr := ethcrypto.PubkeyToAddress(k.PublicKey)
			account := statedb.CreateAccount(addr)
			account.SetBalance(big.NewInt(10000))
		}
		root, _ = statedb.Commit(false)
		statedb, _ = statedb.New(root)
	}
	return root, time.Now().Sub(commitBegin)
}

func PreparePrivateKeys(round int, num int64) []*ecdsa.PrivateKey {
	privkeys := make([]*ecdsa.PrivateKey, 0, NumAccounts)
	keyInt := big.NewInt(0)
	keyInt.SetBytes(common.Hex2Bytes(InitPrivateKeyHex))
	keyInt.Add(keyInt, big.NewInt(int64(round)*num))
	for i := int64(1); i <= num; i++ {
		keyInt.Add(keyInt, big.NewInt(1))
		privkeys = append(privkeys, ethcrypto.ToECDSA(keyInt.Bytes()))
	}
	return privkeys
}
