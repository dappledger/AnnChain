package start

import (
	"fmt"

	"go.uber.org/zap"

	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

func Gen_validator(logger *zap.Logger, conf config.Config) {
	privValidator := at.GenPrivValidator(logger)
	privValidatorJSONBytes := wire.JSONBytesPretty(privValidator)
	fmt.Printf(`Generated a new validator!
Paste the following JSON into your %v file

%v

`,
		conf.GetString("priv_validator_file"),
		string(privValidatorJSONBytes),
	)
}
