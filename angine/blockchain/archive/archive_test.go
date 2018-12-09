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

package archive

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/dappledger/AnnChain/module/lib/go-db"
)

func TestArchive(t *testing.T) {
	Threshold := 20
	arch := NewArchive(db.LevelDBBackendStr, "./", Threshold)
	defer func() {
		os.RemoveAll("archive.db")
		arch.db.Close()
	}()

	fileHash1 := "fileHash_1"
	arch.AddItem("1_"+strconv.Itoa(Threshold), fileHash1)
	fileHash2 := "fileHash_2"
	arch.AddItem(strconv.Itoa(Threshold+1)+"_"+strconv.Itoa(2*Threshold), fileHash2)
	fileHash3 := "fileHash_3"
	arch.AddItem(strconv.Itoa(2*Threshold+1)+"_"+strconv.Itoa(3*Threshold), fileHash3)

	ret := arch.QueryFileHash(1)
	assert.Equal(t, ret, []byte(fileHash1))
	ret = arch.QueryFileHash(Threshold)
	assert.Equal(t, ret, []byte(fileHash1))
	ret = arch.QueryFileHash(Threshold + 1)
	assert.Equal(t, ret, []byte(fileHash2))
	ret = arch.QueryFileHash(2 * Threshold)
	assert.Equal(t, ret, []byte(fileHash2))
	ret = arch.QueryFileHash(2*Threshold + 1)
	assert.Equal(t, ret, []byte(fileHash3))
	ret = arch.QueryFileHash(3 * Threshold)
	assert.Equal(t, ret, []byte(fileHash3))
}
