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

package p2p

import (
	"github.com/spf13/viper"
)

const (
	// Switch config keys
	configKeyDialTimeoutSeconds      = "dial_timeout_seconds"
	configKeyHandshakeTimeoutSeconds = "handshake_timeout_seconds"
	configKeyMaxNumPeers             = "max_num_peers"
	configKeyAuthEnc                 = "authenticated_encryption"

	// MConnection config keys
	configKeySendRate = "send_rate"
	configKeyRecvRate = "recv_rate"

	// Fuzz params
	configFuzzEnable               = "fuzz_enable" // use the fuzz wrapped conn
	configFuzzActive               = "fuzz_active" // toggle fuzzing
	configFuzzMode                 = "fuzz_mode"   // eg. drop, delay
	configFuzzMaxDelayMilliseconds = "fuzz_max_delay_milliseconds"
	configFuzzProbDropRW           = "fuzz_prob_drop_rw"
	configFuzzProbDropConn         = "fuzz_prob_drop_conn"
	configFuzzProbSleep            = "fuzz_prob_sleep"
)

func setConfigDefaults(config *viper.Viper) {
	// Switch default config
	config.SetDefault(configKeyDialTimeoutSeconds, 3)
	config.SetDefault(configKeyHandshakeTimeoutSeconds, 20)
	config.SetDefault(configKeyMaxNumPeers, 50)
	config.SetDefault(configKeyAuthEnc, true)

	// MConnection default config
	config.SetDefault(configKeySendRate, 5120000) // 5000KB/s
	config.SetDefault(configKeyRecvRate, 5120000) // 5000KB/s

	// Fuzz defaults
	config.SetDefault(configFuzzEnable, false)
	config.SetDefault(configFuzzActive, false)
	config.SetDefault(configFuzzMode, FuzzModeDrop)
	config.SetDefault(configFuzzMaxDelayMilliseconds, 3000)
	config.SetDefault(configFuzzProbDropRW, 0.2)
	config.SetDefault(configFuzzProbDropConn, 0.00)
	config.SetDefault(configFuzzProbSleep, 0.00)

	config.SetDefault("connection_reset_wait", 300)
}
