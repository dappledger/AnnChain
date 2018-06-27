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

package config

const CONFIGTPL = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

environment = "production"
node_laddr = "tcp://0.0.0.0:46656"
rpc_laddr = "tcp://0.0.0.0:46657"
moniker = "__MONIKER__"
fast_sync = true
db_backend = "leveldb"
seeds = ""
signbyCA = ""

`

//	"auth_by_ca":false
const MYCONFIGTPL = `
{
	"base_fee":0,
	"base_reserve":0,
	"max_txset_size":2000,
	"block_size":2000,
	"block_part_size":65536,
	"disable_data_hash":false,
	"timeout_propose":3000,
	"timeout_propose_delta":500,
	"timeout_prevote":1000,
	"timeout_prevote_delta":500,
	"timeout_precommit":1000,
	"timeout_precommit_delta":500,
	"timeout_commit":1000,
	"skip_timeout_commit":false
}
`
