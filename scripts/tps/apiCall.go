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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"unsafe"

	"github.com/gin-gonic/gin/json"
)

func contractExecute(privkey, contract, abijson, callfunc string, args []interface{}, nonce uint64, commit bool) error {
	req := make(map[string]interface{})
	req["contract"] = contract
	req["privkey"] = privkey
	req["abiDefinition"] = abijson
	req["method"] = callfunc
	req["params"] = args
	req["nonce"] = nonce
	bytesData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	reader := bytes.NewReader(bytesData)
	if !commit {
		url := "http://0.0.0.0:8889/contract/call/async"

		request, err := http.Post(url, "application/json;charset=UTF-8", reader)
		if err != nil {
			fmt.Println(err.Error())
		}

		respBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println(err.Error())
		}

		str := (*string)(unsafe.Pointer(&respBytes))
		fmt.Println(*str)

		return nil
	}

	url := "http://0.0.0.0:8889/contract/call"
	request, err := http.Post(url, "application/json;charset=UTF-8", reader)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	str := (*string)(unsafe.Pointer(&respBytes))
	fmt.Println(*str)

	return nil
}
