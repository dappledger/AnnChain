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

environment = "development"
proxy_app = "tcp://127.0.0.1:46658"
moniker = "__MONIKER__"
node_laddr = "tcp://0.0.0.0:46656"
seeds = ""
fast_sync = true
db_backend = "leveldb"
rpc_laddr = "tcp://0.0.0.0:46657"
api_laddr = "tcp://0.0.0.0:46659"
signbyCA = ""

#log_level:
	# -1 DebugLevel logs are typically voluminous, and are usually disabled in production.
	#  0 InfoLevel is the default logging priority.
	#  1 WarnLevel logs are more important than Info, but don't need individual human review.
	#  2 ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs.
	#  3 DPanicLevel logs are particularly important errors. In development the logger panics after writing the message.
	#  4 PanicLevel logs a message, then panics.
	#  5 FatalLevel logs a message, then calls os.Exit(1)
`
