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
	"io/ioutil"

	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	"golang.org/x/crypto/openpgp/armor"
)

func EncodeArmor(blockType string, headers map[string]string, data []byte) string {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, blockType, headers)
	if err != nil {
		PanicSanity("Error encoding ascii armor: " + err.Error())
	}
	_, err = w.Write(data)
	if err != nil {
		PanicSanity("Error encoding ascii armor: " + err.Error())
	}
	err = w.Close()
	if err != nil {
		PanicSanity("Error encoding ascii armor: " + err.Error())
	}
	return string(buf.Bytes())
}

func DecodeArmor(armorStr string) (blockType string, headers map[string]string, data []byte, err error) {
	buf := bytes.NewBufferString(armorStr)
	block, err := armor.Decode(buf)
	if err != nil {
		return "", nil, nil, err
	}
	data, err = ioutil.ReadAll(block.Body)
	if err != nil {
		return "", nil, nil, err
	}
	return block.Type, block.Header, data, nil
}
