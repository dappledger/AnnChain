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
	"bytes"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSimple(t *testing.T) {

	MixEntropy([]byte("someentropy"))

	plaintext := []byte("sometext")
	secret := []byte("somesecretoflengththirtytwo===32")
	ciphertext := EncryptSymmetric(plaintext, secret)

	plaintext2, err := DecryptSymmetric(ciphertext, secret)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(plaintext, plaintext2) {
		t.Errorf("Decrypted plaintext was %X, expected %X", plaintext2, plaintext)
	}

}

func TestSimpleWithKDF(t *testing.T) {

	MixEntropy([]byte("someentropy"))

	plaintext := []byte("sometext")
	secretPass := []byte("somesecret")
	secret, err := bcrypt.GenerateFromPassword(secretPass, 12)
	if err != nil {
		t.Error(err)
	}
	secret = Sha256(secret)

	ciphertext := EncryptSymmetric(plaintext, secret)

	plaintext2, err := DecryptSymmetric(ciphertext, secret)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(plaintext, plaintext2) {
		t.Errorf("Decrypted plaintext was %X, expected %X", plaintext2, plaintext)
	}

}
