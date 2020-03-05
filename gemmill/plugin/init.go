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

package plugin

import (
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	dbm "github.com/dappledger/AnnChain/gemmill/modules/go-db"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/refuse_list"
	"github.com/dappledger/AnnChain/gemmill/types"
)

const (
	PluginNoncePrefix = "pn-"
)

var (
	pluginTypeAdminOP    = byte(0x01)
	pluginTypeSuspect    = byte(0x02)
	pluginTypeQueryCache = byte(0x03)
)

type (
	InitParams struct {
		Logger     *zap.Logger
		StateDB    dbm.DB
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKey
		RefuseList *refuse_list.RefuseList
		Validators **types.ValidatorSet
	}

	ReloadParams struct {
		Logger     *zap.Logger
		StateDB    dbm.DB
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKey
		RefuseList *refuse_list.RefuseList
		Validators **types.ValidatorSet
	}

	BeginBlockParams struct {
		Block *types.Block
	}

	BeginBlockReturns struct {
	}

	ExecBlockParams struct {
		Block       *types.Block
		EventSwitch types.EventSwitch
		EventCache  types.EventCache
		ValidTxs    types.Txs
		InvalidTxs  []types.ExecuteInvalidTx
	}

	ExecBlockReturns struct {
	}

	EndBlockParams struct {
		Block             *types.Block
		ChangedValidators []*types.ValidatorAttr
		NextValidatorSet  *types.ValidatorSet
	}

	EndBlockReturns struct {
		NextValidatorSet *types.ValidatorSet
	}

	// IPlugin defines the behavior of the core plugins
	IPlugin interface {
		types.Eventable

		// DeliverTx return false means the tx won't be pass on to proxy app
		DeliverTx(tx []byte, i int) (bool, error)

		// CheckTx return false means the tx won't be pass on to proxy app
		CheckTx(tx []byte) (bool, error)

		// BeginBlock just mock the abci Blockaware interface
		BeginBlock(*BeginBlockParams) (*BeginBlockReturns, error)

		// ExecBlock receives block
		ExecBlock(*ExecBlockParams) (*ExecBlockReturns, error)

		// EndBlock just mock the abci Blockaware interface
		EndBlock(*EndBlockParams) (*EndBlockReturns, error)

		// Reset is called when u don't need to maintain the plugin status
		Reset()

		// InitPlugin custom the initialization of the plugin
		Init(*InitParams)

		// Reload reloads private fields of the plugin
		Reload(*ReloadParams)

		Stop()
	}
)

func init() {
	_ = wire.RegisterInterface(
		struct{ IPlugin }{},
		wire.ConcreteType{&AdminOp{}, pluginTypeAdminOP},
	)
}
