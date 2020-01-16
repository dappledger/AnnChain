// Copyright © 2017 ZhongAn Technology
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"

	"github.com/dappledger/AnnChain/chain/commands/global"
	"github.com/spf13/cobra"
)

const (
	cmdName        = "genesis"
	cmdDescription = "commands for genesis"
)

var (
	rootCmd = &cobra.Command{
		Use:        cmdName,
		Short:      cmdDescription,
		SuggestFor: []string{"genesis"},
	}
)

func init() {
	globalFlags := global.GFlags()
	rootCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "enable client-side debug logging")
	rootCmd.PersistentFlags().StringVar(&globalFlags.RuntimeDir, "runtime", global.DefaultRuntimeDir, "runtime dir")
	rootCmd.PersistentFlags().StringVar(&globalFlags.LogDir, "log_path", "", "log path, default: ./output.log")
	rootCmd.PersistentFlags().StringVar(&globalFlags.AuditLogDir, "audit_log_path", "", "audit log path, default: ./audit.log")
	rootCmd.PersistentFlags().StringVar(&globalFlags.ApiAddr, "apiaddr", global.DefaultApiAddr, "api listening address")

	rootCmd.AddCommand(
		NewInitCommand(),
		NewRunCommand(),
		NewShowCommand(),
		NewVersionCommand(),
		NewResetCommand(),
	)

	cobra.EnablePrefixMatching = true
}

func Start() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("execute error: ", err)
		os.Exit(-1)
	}
}
