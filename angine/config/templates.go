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

import "github.com/spf13/viper"

//const CONFIGTPL = `# This is a TOML config file.
//# For more information, see https://github.com/toml-lang/toml
//
//environment = "development"
//moniker = "__MONIKER__"
//db_backend = "leveldb"
//
//p2p_laddr = "tcp://0.0.0.0:46656"           # p2p port that this node is listening
//rpc_laddr = "tcp://0.0.0.0:46657"           # rpc port this node is exposing
//event_laddr = "tcp://0.0.0.0:46658"         # chorus uses a exposed port for events function
//seeds = ""
//
//auth_by_ca = true                           # auth by ca general switch
//non_validator_auth_by_ca = true             # whether non-validator nodes need auth by ca,
//                                            # only effective when auth_by_ca is true
//signbyCA = ""                               # auth signature from CA
//
//fast_sync = true
//skip_upnp = true
//
//log_path = ""
//#log_level:
//	# -1 DebugLevel logs are typically voluminous, and are usually disabled in production.
//	#  0 InfoLevel is the default logging priority.
//	#  1 WarnLevel logs are more important than Info, but don't need individual human review.
//	#  2 ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs.
//	#  3 DPanicLevel logs are particularly important errors. In development the logger panics after writing the message.
//	#  4 PanicLevel logs a message, then panics.
//	#  5 FatalLevel logs a message, then calls os.Exit(1)
//`

func DefaultConfig() (conf *viper.Viper) {
	conf = viper.New()
	conf.Set("environment", "development")
	conf.Set("moniker", "__MONIKER__")
	conf.Set("db_backend", "leveldb")
	conf.Set("moniker", "anonymous")
	conf.Set("p2p_laddr", "tcp://0.0.0.0:46656")
	conf.Set("rpc_laddr", "tcp://0.0.0.0:46657")
	conf.Set("event_laddr", "tcp://0.0.0.0:46658")
	conf.Set("seeds", "")
	conf.Set("auth_by_ca", false)
	conf.Set("non_validator_auth_by_ca", true)
	conf.Set("signbyCA", "")
	conf.Set("fast_sync", true)
	conf.Set("skip_upnp", true)
	conf.Set("log_path", "")
	conf.Set("fast_sync", "true")
	conf.Set("skip_upnp", "true")

	return conf
}
