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

package state

import (
	. "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/xlib/def"
)

type (
	ErrInvalidBlock error
	ErrProxyAppConn error

	ErrUnknownBlock struct {
		Height def.INT
	}

	ErrBlockHashMismatch struct {
		CoreHash []byte
		AppHash  []byte
		Height   def.INT
	}

	ErrAppBlockHeightTooHigh struct {
		CoreHeight def.INT
		AppHeight  def.INT
	}

	ErrLastStateMismatch struct {
		Height def.INT
		Core   []byte
		App    []byte
	}

	ErrStateMismatch struct {
		Got      *State
		Expected *State
	}
)

func (e ErrUnknownBlock) Error() string {
	return Fmt("Could not find block #%d", e.Height)
}

func (e ErrBlockHashMismatch) Error() string {
	return Fmt("App block hash (%X) does not match core block hash (%X) for height %d", e.AppHash, e.CoreHash, e.Height)
}

func (e ErrAppBlockHeightTooHigh) Error() string {
	return Fmt("App block height (%d) is higher than core (%d)", e.AppHeight, e.CoreHeight)
}
func (e ErrLastStateMismatch) Error() string {
	return Fmt("Latest block (%d) LastAppHash (%X) does not match app's AppHash (%X)", e.Height, e.Core, e.App)
}

func (e ErrStateMismatch) Error() string {
	return Fmt("State after replay does not match saved state. Got ----\n%v\nExpected ----\n%v\n", e.Got, e.Expected)
}
