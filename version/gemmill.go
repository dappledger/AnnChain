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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dappledger/AnnChain/gemmill"
)

func getCommitHash() (hash string, err error) {

	goPath := os.Getenv("GOPATH")
	paths := strings.Split(goPath, ":")

	var bytez []byte
	for _, v := range paths {
		rootPwd := filepath.Join(v, "src", "github.com/dappledger/AnnChain/gemmill")
		bytez, err = ioutil.ReadFile(filepath.Join(rootPwd, ".git/HEAD"))
		if err != nil {
			continue
		}

		hash = string(bytez)
		prefix := hash[:4]
		if prefix == "ref:" {
			file := filepath.Join(rootPwd, ".git", hash[5:len(hash)-1])
			bytez, err = ioutil.ReadFile(file)
			if err != nil {
				break
			}
			hash = string(bytez)
		}
		return
	}
	return
}

func main() {

	hash, _ := getCommitHash()
	if len(hash) > 8 {
		hash = hash[:8]
	}
	fmt.Printf("%s-%s\n", gemmill.Gversion, hash)
}
