/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package cmd

import (
	"fmt"
	"os"

	acfg "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/src/chain/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset PrivValidator, clean the data and shards",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		angineconf := acfg.GetConfig(viper.GetString("runtime"))
		conf := config.GetConfig(angineconf)
		os.RemoveAll(conf.GetString("db_dir"))
		os.RemoveAll(conf.GetString("shards"))
		resetPrivValidator(angineconf.GetString("priv_validator_file"))
	},
}

func init() {
	RootCmd.AddCommand(resetCmd)
}

func resetPrivValidator(privValidatorFile string) {
	var (
		privValidator *types.PrivValidator
		logger        *zap.Logger
	)

	if _, err := os.Stat(privValidatorFile); err == nil {
		privValidator = types.LoadPrivValidator(logger, privValidatorFile)
		privValidator.Reset()
		fmt.Println("Reset PrivValidator", "file", privValidatorFile)
	} else {
		privValidator = types.GenPrivValidator(logger, nil)
		privValidator.SetFile(privValidatorFile)
		privValidator.Save()
		fmt.Println("Generated PrivValidator", "file", privValidatorFile)
	}
}
