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
	"fmt"

	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
)

type Archive struct {
	db        dbm.DB
	Threshold int64 // height int64
	Client    ArchiveClient
}

type ArchiveClient interface {
	UploadFile(filepath string) (fileHash string, err error)
	DownloadFile(fileHash, filepath string) (err error)
}

var dbName = "archive"

func NewArchive(dbBackend, dbDir string, threshold int64) *Archive {
	archiveDB := dbm.NewDB(dbName, dbBackend, dbDir)
	return &Archive{
		db:        archiveDB,
		Threshold: threshold,
		// client implement ArchiveClient
		//	Client:    client,
	}
}

func (ar *Archive) QueryFileHash(height int64) (ret []byte) {
	origin := (height-1)/ar.Threshold*ar.Threshold + 1
	key := fmt.Sprintf("%d_%d", origin, origin-1+ar.Threshold)
	ret = ar.db.Get([]byte(key))
	return
}

func (ar *Archive) AddItem(originHeight int64, value string) {
	key := fmt.Sprintf("%d_%d", originHeight+1, originHeight+ar.Threshold)
	ar.db.SetSync([]byte(key), []byte(value))
}

func (ar *Archive) Close() {
	ar.db.Close()
}
