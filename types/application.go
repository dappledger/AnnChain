package types

import (
	"gitlab.zhonganonline.com/ann/ann-module/lib/go-config"
)

type Application interface {
	GetEngineHooks() Hooks
	CompatibleWithAngine()
	CheckTx([]byte) error
	Query([]byte) Result
	Info() ResultInfo
	Start()
	Stop()
}

type AppMaker func(config.Config) Application
