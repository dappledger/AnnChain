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

package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego"
	_ "github.com/mattn/go-sqlite3"

	"github.com/dappledger/AnnChain/chain/commands/global"
	"github.com/dappledger/AnnChain/chain/core"
	"github.com/dappledger/AnnChain/gemmill/config"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-utils"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/types"
)

type KeyInfo struct {
	Privkey string
	Pubkey  string
	Address string
}

func GenKeyInfo(keyType string) (string, error) {
	privkey, err := crypto.GenPrivkeyByType(keyType)
	if err != nil {
		return "", err
	}
	return genKeyInfo(privkey)
}

func genKeyInfo(privkey crypto.PrivKey) (string, error) {
	pubkey := privkey.PubKey()
	var info KeyInfo
	info.Privkey = privkey.KeyString()
	info.Pubkey = pubkey.KeyString()
	info.Address = fmt.Sprintf("%X", pubkey.Address())
	str, err := json.Marshal(&info)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

type NodeInit struct {
	ConfigPath  string `form:"config_path"`
	Chainid     string `form:"chainid"`
	AppName     string `form:"app_list"`
	CryptoType  string `form:"crypto_list"`
	PeerPrivkey string `form:"peer_privkey"`
	P2PPort     string `form:"p2p_port"`
	RpcPort     string `form:"rpc_port"`
	Peers       string `form:"peers"`
	AuthPrivkey string `form:"auth_privkey"`
	LogPath     string `form:"log_dir"`
	Env         string `form:"environment_list"`
	Genesisfile string `form:"genesisfile"`

	KeyInfo string `form:"keyinfo"`
}

func (n *NodeInit) CheckData() error {
	if len(n.P2PPort) == 0 {
		n.P2PPort = "46656"
	} else {
		if !utils.CheckNumber(n.P2PPort) {
			return errors.New("p2p port should be a number")
		}
	}
	if len(n.RpcPort) == 0 {
		n.RpcPort = "46657"
	} else {
		if !utils.CheckNumber(n.RpcPort) {
			return errors.New("rpc port should be a number")
		}
	}
	if len(n.AppName) == 0 {
		n.AppName = "evm"
	}
	if len(n.Peers) != 0 && !utils.CheckIPAddrSlc(n.Peers) {
		return errors.New("peers should be <addr1>,<addr2>")
	}
	if len(n.CryptoType) == 0 {
		n.CryptoType = crypto.CryptoTypeZhongAn
	}
	if !global.CheckAppName(n.AppName) {
		return errors.New("app name not found")
	}
	if !global.CheckCryptoType(n.CryptoType) {
		return errors.New("crypto type not found")
	}
	return nil
}

func (n *NodeInit) DoInit() error {
	var err error
	if err = n.CheckData(); err != nil {
		return err
	}
	if exist, _ := utils.PathExists(n.ConfigPath); exist {
		return errors.New("path already exist")
	}

	// gen privkey of peer
	privkey := n.PeerPrivkey
	if len(privkey) == 0 && len(n.KeyInfo) != 0 {
		var ki KeyInfo
		if err = json.Unmarshal([]byte(n.KeyInfo), &ki); err != nil {
			return err
		}
		privkey = ki.Privkey
	}
	var pk crypto.PrivKey
	if len(privkey) == 0 {
		n.KeyInfo, err = GenKeyInfo(n.CryptoType)
		if err != nil {
			return err
		}
	} else {
		bytes, err := hex.DecodeString(privkey)
		if err != nil {
			return errors.New(fmt.Sprintf("privkey should be hexadecimal: %v", err))
		}
		pk, err = crypto.GenPrivkeyByBytes(n.CryptoType, bytes)
		if err != nil {
			return err
		}
	}

	// gen chainid
	if len(n.Chainid) == 0 {
		n.Chainid = config.GenChainID()
	}

	// gen ca TODO don't ignore err?
	pubkey, _ := hex.DecodeString(pk.PubKey().KeyString())
	authPrivBytes, _ := hex.DecodeString(n.AuthPrivkey)
	authPk, _ := crypto.GenPrivkeyByBytes(n.CryptoType, authPrivBytes)
	ca := types.SignCA(authPk, pubkey)

	// set tunes config
	tunesConf := global.GenConf()
	tunesConf.Set("p2p_laddr", fmt.Sprintf("tcp://0.0.0.0:%v", n.P2PPort))
	tunesConf.Set("rpc_laddr", fmt.Sprintf("tcp://0.0.0.0:%v", n.RpcPort))
	tunesConf.Set("log_dir", n.LogPath)
	tunesConf.Set("seeds", n.Peers)
	tunesConf.Set("app_name", n.AppName)
	if len(ca) != 0 {
		tunesConf.Set("signbyCA", ca)
		tunesConf.Set("auth_by_ca", true)
	}
	tunesConf.Set("gen_privkey", pk)
	if len(n.CryptoType) != 0 {
		tunesConf.Set("crypto_type", n.CryptoType)
	}
	if len(n.Env) > 0 {
		tunesConf.Set("environment", n.Env)
	}
	if len(n.Genesisfile) > 0 {
		tunesConf.Set("genesis_json_file", n.Genesisfile)
	}

	if err = config.InitRuntime(n.ConfigPath, n.Chainid, tunesConf); err != nil {
		os.RemoveAll(n.ConfigPath)
	}
	return err
}

func DoInitNode(c *beego.Controller) {
	n := &NodeInit{}
	c.ParseForm(n)
	if err := n.DoInit(); err != nil {
		c.Data["json"] = err.Error()
		return
	}

	if _, ok := c.Data["json"]; !ok {
		c.Data["json"] = "Done!"
	}
}

func RunNode(c *beego.Controller) {
	n := &NodeInit{}
	c.ParseForm(n)
	err := global.CheckAndReadRuntimeConfig(n.ConfigPath)
	if err != nil {
		c.Data["json"] = err.Error()
		return
	}
	chret := make(chan error, 0)
	go func() {
		defer log.DumpStack()
		if err := core.RunNodeRet(global.GConf(), "", global.GConf().GetString("app_name")); err != nil {
			chret <- err
		}
	}()
	timer := time.NewTimer(time.Second * 2)
	select {
	case <-timer.C:
		c.Data["json"] = "node is running..."
	case err := <-chret:
		c.Data["json"] = fmt.Sprintf("Failed to start node: %v", err)
	}
}

func CloseServer(c *beego.Controller) {
	beego.BeeApp.Server.Close()
}
