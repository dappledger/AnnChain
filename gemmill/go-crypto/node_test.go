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

package crypto

import (
	"math/rand"
	"testing"
	"time"
)

var (
	signdata = []byte("this data just test for sign/verify!!!")
	localkey = []byte("1234567890abcdef")
	tlen     = 10 * 1024
	testdata []byte

	rd *rand.Rand
)

func init() {
	rd = rand.New(rand.NewSource(time.Now().UnixNano()))

	testdata = make([]byte, tlen)

	for i := 0; i < tlen; i++ {
		testdata[i] = byte(rd.Int31n(127))
	}
}

func int2byte(v int, buf []byte) {
	buf[3] = byte(v & 0xff)
	buf[2] = byte(v >> 8 & 0xff)
	buf[1] = byte(v >> 16 & 0xff)
	buf[0] = byte(v >> 24 & 0xff)
}

func byte2int(buf []byte) int {
	return int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3])
}

//------------------------------------ed25519---------------------------------------
func BenchmarkSignVerify_ed25519(tb *testing.B) {
	node_crypto_type = "ed25519"

	for i := 0; i < tb.N; i++ {
		if !signVerify() {
			tb.Fatal("signVerify failed")
		}
	}
}

func TestSign(t *testing.T) {
	node_crypto_type = "ed25519"
	if !signVerify() {
		t.Fatal("signVerify(ed25519) failed")
	}
}

func signVerify() bool {
	priv := GenNodePrivKey()
	pub := priv.PubKey()
	data := signdata

	sig := priv.Sign(data)

	return pub.VerifyBytes(data, sig)
}
