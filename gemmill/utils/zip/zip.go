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
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CompressDir(dir string) (err error) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	fzip, err := os.Create(dir + ".zip")
	if err != nil {
		return
	}
	defer fzip.Close()
	w := zip.NewWriter(fzip)
	defer w.Close()
	for _, file := range fs {
		fw, errC := w.Create(file.Name())
		if errC != nil {
			err = errC
			return
		}
		fileContent, errR := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if errR != nil {
			err = errR
			return
		}
		_, err = fw.Write(fileContent)
		if err != nil {
			return
		}
	}
	return
}

func Decompress(file, dir string) (err error) {
	err = os.Mkdir(dir, 0777)
	if err != nil {
		return
	}
	cf, err := zip.OpenReader(file)
	if err != nil {
		return
	}
	defer cf.Close()
	for _, file := range cf.File {
		rc, errO := file.Open()
		if errO != nil {
			err = errO
			return
		}
		defer rc.Close()
		f, errC := os.Create(filepath.Join(dir, file.Name))
		if errC != nil {
			err = errC
			return
		}
		defer f.Close()
		_, err = io.Copy(f, rc)
		if err != nil {
			return
		}
	}
	return
}
