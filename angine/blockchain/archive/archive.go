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
	"strconv"

	dbm "github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

type Archive struct {
	db        dbm.DB
	Threshold def.INT
	Client    ArchiveClient
}

type ArchiveClient interface {
	UploadFile(filepath string) (fileHash string, err error)
	DownloadFile(fileHash, filepath string) (err error)
}

var dbName = "archive"

func NewArchive(dbBackend, dbDir string, threshold def.INT) *Archive {
	archiveDB := dbm.NewDB(dbName, dbBackend, dbDir)
	return &Archive{
		db:        archiveDB,
		Threshold: threshold,
	}
}

func (ar *Archive) QueryFileHash(height def.INT) (ret []byte) {
	origin := (height-1)/ar.Threshold*ar.Threshold + 1
	key := strconv.FormatInt(origin, 10) + "_" + strconv.FormatInt(origin-1+ar.Threshold, 10)
	ret = ar.db.Get([]byte(key))
	return
}

func (ar *Archive) AddItem(key, value string) {
	ar.db.SetSync([]byte(key), []byte(value))
}

func (ar *Archive) Close() {
	ar.db.Close()
}
