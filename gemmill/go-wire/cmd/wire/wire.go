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
	"os"
	"strings"

	"github.com/dappledger/AnnChain/gemmill/go-wire/expr"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
)

func main() {
	input := ""
	if len(os.Args) > 2 {
		input = strings.Join(os.Args[1:], " ")
	} else if len(os.Args) == 2 {
		input = os.Args[1]
	} else {
		fmt.Println("Usage: wire 'u64:1 u64:2 <sig:Alice>'")
		return
	}

	// fmt.Println(input)
	got, err := expr.ParseReader(input, strings.NewReader(input))
	if err != nil {
		Exit("Error parsing input: " + err.Error())
	}
	gotBytes, err := got.(expr.Byteful).Bytes()
	if err != nil {
		Exit("Error serializing parsed input: " + err.Error())
	}

	fmt.Println(gcmn.Fmt("%X", gotBytes))
}

func Exit(s string) {
	fmt.Printf(s + "\n")
	os.Exit(1)
}
