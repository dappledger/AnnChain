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

package pbft

import (
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
)

// kind of arbitrary
var Spec = "1"     // async
var Major = "0"    //
var Minor = "2"    // replay refactor
var Revision = "2" // validation -> commit

var Version = gcmn.Fmt("v%s/%s.%s.%s", Spec, Major, Minor, Revision)
