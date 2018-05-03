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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/src/chain/node"
)

// nodeCmd represents the node command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a blockchain full-capacity node",
	Long:  ``,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		node.RunNode(viper.GetViper())
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("chain_id", "", "", "specify the chain id when the node is joining it without genesis file")
	runCmd.Flags().BoolP("pprof", "", false, "start golang profile at port :6060")
	runCmd.Flags().BoolP("test", "", false, "run the node in test mode")

	viper.BindPFlag("chain_id", runCmd.Flag("chain_id"))
	viper.BindPFlag("pprof", runCmd.Flag("pprof"))
	viper.BindPFlag("test", runCmd.Flag("test"))
}
