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


package evm

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	ethcmn "github.com/dappledger/AnnChain/eth/common"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/crypto"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/angine/types"
)

const (
	txCount = 10000
)

func TestExe(t *testing.T) {
	validateRoutineCount = 8

	fmt.Println("CPU:", validateRoutineCount)
	fmt.Println("making random txs...")
	txs := randomTxs()

	fmt.Println("begin test...")

	err := RunCPUSerialVerifyTest(txs)
	if err != nil {
		t.Error(err)
	}

	err = RunCPUParallelVerifyTest(txs)
	if err != nil {
		t.Error(err)
	}

	err = RunGPUVerifyTest(txs)
	if err != nil {
		t.Error(err)
	}
}

func RunCPUSerialVerifyTest(txs [][]byte) error {
	begin := time.Now()

	exeWithCPUSerialVeirfy(EthSigner, txs, nil, func(index int, raw []byte, tx *ethtypes.Transaction, abi []byte) {
		if tx != nil {
			_, err := ethtypes.Sender(EthSigner, tx)
			panicErr(err)
		}
	}, func(raw []byte, err error) {
		fmt.Println(err)
	})

	fmt.Println("serial   use time", time.Now().Sub(begin).Seconds())

	return nil
}

func RunCPUParallelVerifyTest(txs [][]byte) error {
	begin := time.Now()

	exeWithCPUParallelVeirfy(EthSigner, txs, nil, func(index int, raw []byte, tx *ethtypes.Transaction, abi []byte) {
	}, func(raw []byte, err error) {
		fmt.Println(err)
	})

	fmt.Println("parallel use time", time.Now().Sub(begin).Seconds())

	return nil
}

func RunGPUVerifyTest(txs [][]byte) error {
	begin := time.Now()

	// exeWithGPUVeirfy(EthSigner, txs, nil, func(index int, raw []byte, tx *ethtypes.Transaction, abi []byte) {
	// }, func(raw []byte, err error) {
	// 	fmt.Println(err)
	// })

	fmt.Println("gpu     use time", time.Now().Sub(begin).Seconds())

	return nil
}

func randomTxs() (txs [][]byte) {
	privkey := "7d73c3dafd3c0215b8526b26f8dbdb93242fc7dcfbdfa1000d93436d577c3b94"
	for i := 0; i < txCount; i++ {
		tx := ethtypes.NewTransaction(uint64(i), ethcmn.Address{}, big.NewInt(0), big.NewInt(90000000000), big.NewInt(0), []byte("dsfsdflsjflsajfldsjflasjfljflsjflksjflkjioejfoijoijelkfno34534n5,34n5k34n5,34n,mn,ren,mn43kn,mrne,mrn,rwnfrewrne4j443i534h543tnkjrenknd"))
		key, err := crypto.HexToECDSA(privkey)
		panicErr(err)
		sig, err := crypto.Sign(tx.SigHash(EthSigner).Bytes(), key)
		panicErr(err)
		sigTx, err := tx.WithSignature(EthSigner, sig)
		panicErr(err)
		b, err := rlp.EncodeToBytes(sigTx)
		panicErr(err)

		b = types.WrapTx(EVMTxTag, b)

		txs = append(txs, b)
	}

	return
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestNothing(t *testing.T) {
}
