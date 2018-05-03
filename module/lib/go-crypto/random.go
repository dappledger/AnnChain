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
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/hex"
	"io"
	"sync"

	. "github.com/dappledger/AnnChain/module/lib/go-common"
)

var gRandInfo *randInfo

func init() {
	gRandInfo = &randInfo{}
	gRandInfo.MixEntropy(randBytes(32)) // Init
}

// Mix additional bytes of randomness, e.g. from hardware, user-input, etc.
// It is OK to call it multiple times.  It does not diminish security.
func MixEntropy(seedBytes []byte) {
	gRandInfo.MixEntropy(seedBytes)
}

// This only uses the OS's randomness
func randBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		PanicCrisis(err)
	}
	return b
}

// This uses the OS and the Seed(s).
func CRandBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := gRandInfo.Read(b)
	if err != nil {
		PanicCrisis(err)
	}
	return b
}

// RandHex(24) gives 96 bits of randomness, strong enough for most purposes.
func CRandHex(numDigits int) string {
	return hex.EncodeToString(CRandBytes(numDigits / 2))
}

// Returns a crand.Reader mixed with user-supplied entropy
func CReader() io.Reader {
	return gRandInfo
}

//--------------------------------------------------------------------------------

type randInfo struct {
	mtx          sync.Mutex
	seedBytes    [32]byte
	cipherAES256 cipher.Block
	streamAES256 cipher.Stream
	reader       io.Reader
}

// You can call this as many times as you'd like.
// XXX TODO review
func (ri *randInfo) MixEntropy(seedBytes []byte) {
	ri.mtx.Lock()
	defer ri.mtx.Unlock()
	// Make new ri.seedBytes
	hashBytes := Sha256(seedBytes)
	hashBytes32 := [32]byte{}
	copy(hashBytes32[:], hashBytes)
	ri.seedBytes = xorBytes32(ri.seedBytes, hashBytes32)
	// Create new cipher.Block
	var err error
	ri.cipherAES256, err = aes.NewCipher(ri.seedBytes[:])
	if err != nil {
		PanicSanity("Error creating AES256 cipher: " + err.Error())
	}
	// Create new stream
	ri.streamAES256 = cipher.NewCTR(ri.cipherAES256, randBytes(aes.BlockSize))
	// Create new reader
	ri.reader = &cipher.StreamReader{S: ri.streamAES256, R: crand.Reader}
}

func (ri *randInfo) Read(b []byte) (n int, err error) {
	ri.mtx.Lock()
	defer ri.mtx.Unlock()
	return ri.reader.Read(b)
}

func xorBytes32(bytesA [32]byte, bytesB [32]byte) (res [32]byte) {
	for i, b := range bytesA {
		res[i] = b ^ bytesB[i]
	}
	return res
}
