package start

import (
	"fmt"

	"go.uber.org/zap"

	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

func Show_validator(logger *zap.Logger, conf config.Config) {
	privValidatorFile := conf.GetString("priv_validator_file")
	privValidator := at.LoadOrGenPrivValidator(logger, privValidatorFile)
	fmt.Println(string(wire.JSONBytesPretty(privValidator.PubKey)))
}
