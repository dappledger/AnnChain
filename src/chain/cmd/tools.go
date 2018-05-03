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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/angine/blockchain"
	acfg "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/angine/state"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/xlib/def"
	"github.com/dappledger/AnnChain/src/chain/app/evm"
	"github.com/dappledger/AnnChain/src/chain/config"
	"github.com/dappledger/AnnChain/src/chain/node"
)

// for the case of conflicting with names already exist
const (
	REVERT_TO_HEIGHT   = "revert_to_height"
	REVERT_APP_NAME    = "revert_app_name"
	REVERT_BRANCH_NAME = "revert_pre_branch_name"
)

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "branch chain from a history height",
	Long:  ``,
	Args:  cobra.OnlyValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := revertToHeight(); err != nil {
			fmt.Println("branch err:", err)
			return
		}
		fmt.Println("branch ok")
	},
}

func init() {
	branchCmd.Flags().Uint("height", 0, "height of a history to revert")
	branchCmd.Flags().String("appname", "", "app name of data(metropolis,evm,...)")
	branchCmd.Flags().String("pre_branch_name", "", "name of the branch name")
	viper.BindPFlag(REVERT_TO_HEIGHT, branchCmd.Flag("height"))
	viper.BindPFlag(REVERT_APP_NAME, branchCmd.Flag("appname"))
	viper.BindPFlag(REVERT_BRANCH_NAME, branchCmd.Flag("pre_branch_name"))
	RootCmd.AddCommand(branchCmd)
}

type AppToolItfc interface {
	Init(datadir string) error
	LastHeightHash() (def.INT, []byte)
	BackupLastBlock(branchName string) error
	DelBackup(branchName string)
	RevertFromBackup(branchName string) error
	SaveNewLastBlock(fromHeight def.INT, fromAppHash []byte) error
}

type RevertTool struct {
	apptool   AppToolItfc
	blocktool blockchain.StoreTool
	statetool state.StateTool
	privtool  agtypes.PrivValidatorTool
}

const (
	APP_TOOL   = 1
	BLOCK_TOOL = 1 << 1
	STATE_TOOL = 1 << 2
	PRIV_TOOL  = 1 << 3
)

func (rt *RevertTool) Init(appname string, agconf, appconf *viper.Viper) error {
	dbdir := appconf.GetString("db_dir")
	if len(viper.GetString("db_dir")) > 0 {
		dbdir = viper.GetString("db_dir")
		appconf.Set("db_dir", dbdir)
	}
	switch appname {
	case "evm":
		rt.apptool = &evm.AppTool{}
	case "metropolis":
		rt.apptool = &node.AppTool{}
	default:
		return errors.New(fmt.Sprintf("illegal appname %v", appname))
	}
	var err error
	if err = rt.apptool.Init(dbdir); err != nil {
		return err
	}

	if err = rt.blocktool.Init(agconf); err != nil {
		return err
	}

	if err = rt.statetool.Init(agconf); err != nil {
		return err
	}

	if err = rt.privtool.Init(agconf.GetString("priv_validator_file")); err != nil {
		return err
	}
	return nil
}

func (rt *RevertTool) checkDataState(toHeight def.INT) error {
	appHeight, _ := rt.apptool.LastHeightHash()
	storeHeight := def.INT(rt.blocktool.LastHeight())
	stateHeight := def.INT(rt.statetool.LastHeight())
	if storeHeight == 0 || appHeight == 0 || stateHeight == 0 {
		return errors.New(fmt.Sprintf("illegal height 0,appHeight:%v,storeHeight:%v,stateHeight:%v", appHeight, storeHeight, stateHeight))
	}
	if toHeight > appHeight || toHeight > storeHeight || toHeight > stateHeight {
		// TODO do check as state/execution.go
		return errors.New(fmt.Sprintf("illegal height bigger than toHeight:%v,appHeight:%v,storeHeight:%v,stateHeight:%v", toHeight, appHeight, storeHeight, stateHeight))
	}
	return nil
}

func (rt *RevertTool) backupData(branchName string) error {
	var err error
	var failDel int
	if err = rt.apptool.BackupLastBlock(branchName); err != nil {
		return err
	}
	failDel |= APP_TOOL
	if err = rt.blocktool.BackupLastBlock(branchName); err != nil {
		rt.delBackup(branchName, failDel)
		return err
	}
	failDel |= BLOCK_TOOL
	if err = rt.statetool.BackupLastState(branchName); err != nil {
		rt.delBackup(branchName, failDel)
		return err
	}
	failDel |= STATE_TOOL
	if err = rt.privtool.BackupData(branchName); err != nil {
		rt.delBackup(branchName, failDel)
		return err
	}
	return nil
}

func (rt *RevertTool) delBackup(branchName string, failDel int) {
	if (failDel & APP_TOOL) > 0 {
		rt.apptool.DelBackup(branchName)
	}
	if (failDel & BLOCK_TOOL) > 0 {
		rt.blocktool.DelBackup(branchName)
	}
	if (failDel & STATE_TOOL) > 0 {
		rt.statetool.DelBackup(branchName)
	}
	if (failDel & PRIV_TOOL) > 0 {
		rt.privtool.DelBackup(branchName)
	}
}

func (rt *RevertTool) revertFromBackup(branchName string, failDel int) {
	if (failDel & APP_TOOL) > 0 {
		rt.apptool.RevertFromBackup(branchName)
	}
	if (failDel & BLOCK_TOOL) > 0 {
		rt.blocktool.RevertFromBackup(branchName)
	}
	if (failDel & STATE_TOOL) > 0 {
		rt.statetool.RevertFromBackup(branchName)
	}
	if (failDel & PRIV_TOOL) > 0 {
		rt.privtool.RevertFromBackup(branchName)
	}
}

func (rt *RevertTool) saveBranchNew(branchName string, toHeight def.INT) error {
	lastBlock, lastBlockMeta, lastBlockID := rt.blocktool.LoadBlock(toHeight)
	if lastBlock == nil || lastBlockMeta == nil || lastBlockID == nil {
		return errors.New("can't find block of the height, is it a height in the future?")
	}
	var err error
	var failDel int
	if err = rt.apptool.SaveNewLastBlock(toHeight, lastBlock.Header.AppHash); err != nil {
		return err
	}
	failDel |= APP_TOOL
	if err = rt.blocktool.SaveNewLastBlock(toHeight); err != nil {
		rt.revertFromBackup(branchName, failDel)
		return err
	}
	failDel |= BLOCK_TOOL
	if err = rt.statetool.SaveNewState(lastBlock, lastBlockMeta, lastBlockID); err != nil {
		rt.revertFromBackup(branchName, failDel)
		return err
	}
	failDel |= STATE_TOOL
	if err = rt.privtool.SaveNewPrivV(toHeight); err != nil {
		rt.revertFromBackup(branchName, failDel)
		return err
	}
	return nil
}

func revertToHeight() error {
	agconf := acfg.GetConfig(viper.GetString("runtime"))
	appconf := config.GetConfig(agconf)
	toHeight := viper.GetInt64(REVERT_TO_HEIGHT)
	if toHeight <= 0 {
		return errors.New("missing param: height")
	}
	appname := viper.GetString(REVERT_APP_NAME)
	if len(appname) == 0 {
		return errors.New("missing param: appname")
	}
	preBranchName := viper.GetString(REVERT_BRANCH_NAME)
	if len(preBranchName) == 0 {
		return errors.New("missing param: pre_branch_name")
	}

	var rt RevertTool
	err := rt.Init(appname, agconf, appconf)
	if err != nil {
		return err
	}

	if err = rt.checkDataState(toHeight); err != nil {
		return err
	}

	if err = rt.backupData(preBranchName); err != nil {
		return err
	}

	if err = rt.saveBranchNew(preBranchName, toHeight); err != nil {
		return err
	}
	return nil
}
