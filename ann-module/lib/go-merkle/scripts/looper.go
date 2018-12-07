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

package main

import (
	"fmt"

	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	"github.com/dappledger/AnnChain/ann-module/lib/go-db"
	"github.com/dappledger/AnnChain/ann-module/lib/go-merkle"
)

func main() {
	db := db.NewMemDB()
	t := merkle.NewIAVLTree(0, db)
	// 23000ns/op, 43000ops/s
	// for i := 0; i < 10000000; i++ {
	// for i := 0; i < 1000000; i++ {
	for i := 0; i < 1000; i++ {
		t.Set(RandBytes(12), nil)
	}
	t.Save()

	fmt.Println("ok, starting")

	for i := 0; ; i++ {
		key := RandBytes(12)
		t.Set(key, nil)
		t.Remove(key)
		if i%1000 == 0 {
			t.Save()
		}
	}
}
