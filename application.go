package angine

import (
	"gitlab.zhonganonline.com/ann/angine/types"
)

type Application interface {
	GetEngineHooks() types.Hooks
	CompatibleWithAngine()
	CheckTx([]byte) error
	Query([]byte) types.Result
	Info() types.ResultInfo
}
