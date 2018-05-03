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
