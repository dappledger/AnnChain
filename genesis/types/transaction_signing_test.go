// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
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

package types

import (
	"encoding/hex"
	"testing"

	"github.com/dappledger/AnnChain/genesis/eth/rlp"

	ethcmn "github.com/dappledger/AnnChain/genesis/eth/common"
	"github.com/dappledger/AnnChain/genesis/eth/crypto"
)

func TestSha3(t *testing.T) {

	var tx Transaction

	bb, _ := ethcmn.HexToByte("0xf9011b30308e637265617465206163636f756e748e6372656174655f6163636f756e74aa307836353138383435396131646336353938346130633764346133393765643339383665643063383533aa3078373666363463616563666332336135393233656662623566366236393262393161373061613432399e7b227374617274696e6742616c616e6365223a223130303030303030227db884307837323833323338323036613562396132306539633232323637663432323431363834643965386164316435323866333464653237303238373637356363653134316635653730323336336135613765616561653166356330383664643235313539313734663035626435666438356365303766303663356564333630323435393031")

	rlp.DecodeBytes(bb, &tx.Data)

	t.Log(tx.String(), tx.SigHash().Hex())
}

func TestSign(t *testing.T) {

	privKey := "7cb4880c2d4863f88134fd01a250ef6633cc5e01aeba4c862bedbf883a148ba8"

	privateKey := crypto.ToECDSA(ethcmn.Hex2Bytes(privKey))

	bb := crypto.PubkeyToAddress(privateKey.PublicKey)

	sig, err := crypto.Sign([]byte("11111111111111111111111111111111"), privateKey)

	if err != nil {
		t.Error(err)
	}

	t.Log(len(sig), hex.EncodeToString(sig), bb.Hex())
}
