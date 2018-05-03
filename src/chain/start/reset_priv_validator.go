package start

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine/types"
)

// NOTE: this is totally unsafe.
// it's only suitable for testnets.
func Reset_all(logger *zap.Logger, conf *viper.Viper) {
	Reset_priv_validator(logger, conf)
	os.RemoveAll(conf.GetString("db_dir"))
	os.RemoveAll(conf.GetString("cs_wal_dir"))
}

// NOTE: this is totally unsafe.
// it's only suitable for testnets.
func Reset_priv_validator(logger *zap.Logger, conf *viper.Viper) {
	// Get PrivValidator
	var privValidator *types.PrivValidator
	privValidatorFile := conf.GetString("priv_validator_file")
	if _, err := os.Stat(privValidatorFile); err == nil {
		privValidator = types.LoadPrivValidator(logger, privValidatorFile)
		privValidator.Reset()
		fmt.Println("Reset PrivValidator", "file", privValidatorFile)
	} else {
		privValidator = types.GenPrivValidator(logger)
		privValidator.SetFile(privValidatorFile)
		privValidator.Save()
		fmt.Println("Generated PrivValidator", "file", privValidatorFile)
	}
}
