package start

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	at "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/ann-module/lib/go-config"
)

// NOTE: this is totally unsafe.
// it's only suitable for testnets.
func Reset_all(logger *zap.Logger, conf config.Config) {
	Reset_priv_validator(logger, conf)
	os.RemoveAll(conf.GetString("db_dir"))
	os.RemoveAll(conf.GetString("cs_wal_dir"))
}

// NOTE: this is totally unsafe.
// it's only suitable for testnets.
func Reset_priv_validator(logger *zap.Logger, conf config.Config) {
	// Get PrivValidator
	var privValidator *at.PrivValidator
	privValidatorFile := conf.GetString("priv_validator_file")
	if _, err := os.Stat(privValidatorFile); err == nil {
		privValidator = at.LoadPrivValidator(logger, privValidatorFile)
		privValidator.Reset()
		fmt.Println("Reset PrivValidator", "file", privValidatorFile)
	} else {
		privValidator = at.GenPrivValidator(logger)
		privValidator.SetFile(privValidatorFile)
		privValidator.Save()
		fmt.Println("Generated PrivValidator", "file", privValidatorFile)
	}
}
