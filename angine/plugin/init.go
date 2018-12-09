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
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/blockchain/refuse_list"
	sttpb "github.com/dappledger/AnnChain/angine/protos/state"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-db"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
)

const (
	PluginNoncePrefix = "pn-"
)

type (
	InitParams struct {
		Logger     *zap.Logger
		DB         db.DB
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKeyEd25519
		RefuseList *refuse_list.RefuseList
		Validators **agtypes.ValidatorSet
	}

	ReloadParams struct {
		Logger     *zap.Logger
		DB         db.DB
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKeyEd25519
		RefuseList *refuse_list.RefuseList
		Validators **agtypes.ValidatorSet
	}

	BeginBlockParams struct {
		Block *agtypes.BlockCache
	}

	BeginBlockReturns struct {
	}

	ExecBlockParams struct {
		Block       *agtypes.BlockCache
		EventSwitch agtypes.EventSwitch
		EventCache  agtypes.EventCache
		ValidTxs    agtypes.Txs
		InvalidTxs  []agtypes.ExecuteInvalidTx
	}

	ExecBlockReturns struct {
	}

	EndBlockParams struct {
		Block             *agtypes.BlockCache
		ChangedValidators []*agtypes.ValidatorAttr
		NextValidatorSet  *agtypes.ValidatorSet
	}

	EndBlockReturns struct {
		NextValidatorSet *agtypes.ValidatorSet
	}

	// IPlugin defines the behavior of the core plugins
	IPlugin interface {
		agtypes.Eventable

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

func PluginsFromPbData(pbPlugins []*sttpb.Plugin, logger *zap.Logger, statedb db.DB) ([]IPlugin, error) {
	plugins := make([]IPlugin, len(pbPlugins))
	for i := range pbPlugins {
		p, err := PluginFromPbData(pbPlugins[i], logger, statedb)
		if err != nil {
			return nil, err
		}
		plugins[i] = p
	}
	return plugins, nil
}

func PluginFromPbData(pbPlugin *sttpb.Plugin, logger *zap.Logger, statedb db.DB) (IPlugin, error) {
	switch pbPlugin.GetType() {
	case sttpb.Type_PluginSpecialOp:
		var pbSop sttpb.SpecialOp
		if err := agtypes.UnmarshalData(pbPlugin.GetPData(), &pbSop); err != nil {
			return nil, err
		}
		sop := NewSpecialop(logger, statedb)
		if err := json.Unmarshal(pbSop.JSONData, &sop); err != nil {
			return nil, err
		}
		return sop, nil
	case sttpb.Type_PluginSuspect:
	case sttpb.Type_PluginQueryCache:
	default:
		return nil, errors.New("unknow plugin type to unmarshal")
	}
	return nil, nil
}

func ToPbPlugins(plugins []IPlugin) ([]*sttpb.Plugin, error) {
	pbPlugins := make([]*sttpb.Plugin, len(plugins))
	var err error
	for i := range plugins {
		if pbPlugins[i], err = ToPbPlugin(plugins[i]); err != nil {
			return nil, err
		}
	}
	return pbPlugins, err
}

func ToPbPlugin(iplugin IPlugin) (pb *sttpb.Plugin, err error) {
	pb = &sttpb.Plugin{}
	switch pl := iplugin.(type) {
	case *Specialop:
		var jsBys, spOpBys []byte
		if jsBys, err = json.Marshal(pl); err != nil {
			return nil, err
		}
		pb.Type = sttpb.Type_PluginSpecialOp
		var spOp sttpb.SpecialOp
		spOp.JSONData = jsBys
		if spOpBys, err = agtypes.MarshalData(&spOp); err != nil {
			return nil, err
		}
		pb.PData = spOpBys
	case *SuspectPlugin:
	case *QueryCachePlugin:
	default:
		return nil, errors.New("unknown plugin type")
	}
	return
}
