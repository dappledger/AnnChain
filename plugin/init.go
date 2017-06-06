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
