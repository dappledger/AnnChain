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
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show infomation about this node",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(showCmd)

	showCmd.AddCommand([]*cobra.Command{
		&cobra.Command{
			Use:   "pubkey",
			Short: "print this node's public key",
			Long:  "",
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				ang := angine.NewAngine(zap.NewNop(), &angine.Tunes{Runtime: viper.GetString("runtime")})
				cmd.Println(ang.PrivValidator().PubKey)
			},
		},
	}...)
}
