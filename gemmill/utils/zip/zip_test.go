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

package zip

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	dir := "./testZip"
	err := os.Mkdir(dir, 0777)
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	file, err := os.Create(filepath.Join(dir, "text"))
	assert.Nil(t, err)
	defer file.Close()
	fileContent := "just for test"
	reader := bytes.NewReader([]byte(fileContent))
	_, err = io.Copy(file, reader)
	assert.Nil(t, err)

	err = CompressDir(dir)
	assert.Nil(t, err)
	_, err = os.Stat(dir + ".zip")
	assert.Nil(t, err)
	defer os.Remove(dir + ".zip")
	newDir := "./testDecompress"
	err = Decompress(dir+".zip", newDir)
	assert.Nil(t, err)
	defer os.RemoveAll(newDir)
	bytez, err := ioutil.ReadFile(filepath.Join(newDir, "text"))
	assert.Nil(t, err)
	assert.Equal(t, fileContent, string(bytez))
}
