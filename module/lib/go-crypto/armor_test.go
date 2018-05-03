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
)

func TestSimpleArmor(t *testing.T) {
	blockType := "MINT TEST"
	data := []byte("somedata")
	armorStr := EncodeArmor(blockType, nil, data)
	t.Log("Got armor: ", armorStr)

	// Decode armorStr and test for equivalence.
	blockType2, _, data2, err := DecodeArmor(armorStr)
	if err != nil {
		t.Error(err)
	}
	if blockType != blockType2 {
		t.Errorf("Expected block type %v but got %v", blockType, blockType2)
	}
	if !bytes.Equal(data, data2) {
		t.Errorf("Expected data %X but got %X", data2, data)
	}
}
