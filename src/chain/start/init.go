package start

import (
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/angine"
)

func Initfiles(conf *viper.Viper) {
	angine.Initialize(&angine.Tunes{Conf: conf},"")
}