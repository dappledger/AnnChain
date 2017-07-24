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
	"gitlab.zhonganonline.com/ann/angine/refuse_list"
	"gitlab.zhonganonline.com/ann/angine/types"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-crypto"
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-p2p"
)

const (
	PluginNoncePrefix = "pn-"
)

type (
	InitPluginParams struct {
		Switch     *p2p.Switch
		PrivKey    crypto.PrivKeyEd25519
		RefuseList *refuse_list.RefuseList
		Validators **types.ValidatorSet
	}

	BeginBlockParams struct {
		Block *types.Block
	}

	BeginBlockReturns struct {
	}

	EndBlockParams struct {
		Block             *types.Block
		ChangedValidators []*types.ValidatorAttr
		NextValidatorSet  *types.ValidatorSet
	}

	EndBlockReturns struct {
		NextValidatorSet *types.ValidatorSet
	}
)
