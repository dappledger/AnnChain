package start

import (
	"github.com/dappledger/AnnChain/angine"
	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
)

func Initfiles(conf *config.MapConfig) {
	angine.Initialize(&angine.AngineTunes{Conf: conf})
}
