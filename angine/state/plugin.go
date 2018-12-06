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
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
	"github.com/dappledger/AnnChain/angine/plugin"
)

type (

	// Plugin defines the behavior of the core plugins
	IPlugin interface {

		// DeliverTx return false means the tx won't be pass on to proxy app
		DeliverTx(tx []byte, i int) (bool, error)

		// CheckTx return false means the tx won't be pass on to proxy app
		CheckTx(tx []byte) (bool, error)

		// BeginBlock just mock the abci Blockaware interface
		BeginBlock(*plugin.BeginBlockParams) (*plugin.BeginBlockReturns, error)

		// EndBlock just mock the abci Blockaware interface
		EndBlock(*plugin.EndBlockParams) (*plugin.EndBlockReturns, error)

		// Reset is called when u don't need to maintain the plugin status
		Reset()

		// InitPlugin custom the initialization of the plugin
		InitPlugin(*plugin.InitPluginParams)
	}
)

var (
	pluginTypeSpecialOP = byte(0x01)
)

func init() {
	_ = wire.RegisterInterface(
		struct{ IPlugin }{},
		wire.ConcreteType{&plugin.Specialop{}, pluginTypeSpecialOP},
	)
}
