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

package refuse_list

import (
	"os"
	"testing"

	"github.com/dappledger/AnnChain/angine/types"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-db"
)

func TestRefuseList(t *testing.T) {
	refuseList := NewRefuseList(db.LevelDBBackendStr, "./")
	defer func() {
		os.RemoveAll("refuse_list.db")
		refuseList.db.Close()
	}()
	var keyStr = "6FEBD39916627AA0CD7CFDA4A94586F3BA958078621E6E466488A423272B9700"

	pubKey, err := types.StringTo32byte(keyStr)
	assert.Nil(t, err)
	refuseList.AddRefuseKey(pubKey)
	assert.Equal(t, true, refuseList.QueryRefuseKey(pubKey))
	assert.Equal(t, []string{keyStr}, refuseList.ListAllKey())
	refuseList.DeleteRefuseKey(pubKey)
	assert.Equal(t, 0, len(refuseList.ListAllKey()))
}
