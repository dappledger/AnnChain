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

package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/angine"
	agconf "github.com/dappledger/AnnChain/angine/config"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	crypto "github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib"
	"github.com/dappledger/AnnChain/src/chain/node"
	"github.com/spf13/viper"
)

type KeyInfo struct {
	Privkey string
	Pubkey  string
	Address string
}

func GenKeyInfo() string {
	privkey := crypto.GenPrivKeyEd25519()
	return genKeyInfo(&privkey)
}

func genKeyInfo(privkey *crypto.PrivKeyEd25519) string {
	pubkey := privkey.PubKey()
	var info KeyInfo
	info.Privkey = privkey.KeyString()
	info.Pubkey = pubkey.KeyString()
	info.Address = fmt.Sprintf("%X", pubkey.Address())
	str, err := json.Marshal(&info)
	if err != nil {
		return err.Error()
	}
	return string(str)
}

type NodeInit struct {
	ConfigPath  string `form:"config_path"`
	Chainid     string `form:"chainid"`
	PeerPrivkey string `form:"peer_privkey"`
	P2PPort     string `form:"p2p_port"`
	RpcPort     string `form:"rpc_port"`
	EventPort   string `form:"event_port"`
	Peers       string `form:"peers"`
	AuthPrivkey string `form:"auth_privkey"`
	LogPath     string `form:"log_path"`
	Genesisfile string `form:"genesisfile"`

	KeyInfo string `form:"keyinfo"`
}

func (n *NodeInit) CheckData() error {
	if len(n.P2PPort) == 0 {
		n.P2PPort = "46656"
	} else {
		if !xlib.CheckNumber(n.P2PPort) {
			return errors.New("p2p port should be a number")
		}
	}
	if len(n.RpcPort) == 0 {
		n.RpcPort = "46657"
	} else {
		if !xlib.CheckNumber(n.RpcPort) {
			return errors.New("rpc port should be a number")
		}
	}
	if len(n.EventPort) == 0 {
		n.EventPort = "46650"
	} else {
		if !xlib.CheckNumber(n.RpcPort) {
			return errors.New("event port should be a number")
		}
	}
	if len(n.Peers) != 0 && !xlib.CheckIPAddrSlc(n.Peers) {
		return errors.New("peers should be <addr1>,<addr2>...")
	}
	return nil
}

func (n *NodeInit) DoInit(runtime string) error {
	var err error
	if err = n.CheckData(); err != nil {
		return err
	}
	if exist, _ := xlib.PathExists(n.ConfigPath); exist {
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
	var edkey crypto.PrivKeyEd25519
	if len(privkey) == 0 {
		edkey = crypto.GenPrivKeyEd25519()
		n.KeyInfo = genKeyInfo(&edkey)
	} else {
		bytes, err := hex.DecodeString(privkey)
		if err != nil {
			return errors.New(fmt.Sprintf("privkey should be hexadecimal", err))
		}
		copy(edkey[:], bytes)
	}

	// gen chainid
	if len(n.Chainid) == 0 {
		n.Chainid = angine.GenChainid()
	}

	// gen ca
	pubkeyBytes, _ := hex.DecodeString(edkey.PubKey().KeyString())
	plainTxt := append(pubkeyBytes, []byte(n.Chainid)...)
	authPrivBytes, _ := hex.DecodeString(n.AuthPrivkey)
	var ca string
	ca, err = agtypes.SignCA(authPrivBytes, plainTxt)
	if err != nil {
		return err
	}
	fmt.Println("ca:", ca)

	// gen runtime
	if len(n.ConfigPath) > 0 {
		runtime = n.ConfigPath
	}

	// set tunes config
	tunesConf := viper.New()
	tunesConf.Set("p2p_laddr", fmt.Sprintf("tcp://0.0.0.0:%v", n.P2PPort))
	tunesConf.Set("rpc_laddr", fmt.Sprintf("tcp://0.0.0.0:%v", n.RpcPort))
	tunesConf.Set("event_laddr", fmt.Sprintf("tcp://0.0.0.0:%v", n.EventPort))
	tunesConf.Set("seeds", n.Peers)
	if len(n.LogPath) != 0 {
		tunesConf.Set("log_path", n.LogPath)
	}

	if len(ca) != 0 {
		tunesConf.Set("signbyCA", ca)
		tunesConf.Set("auth_by_ca", true)
	}
	if len(n.Genesisfile) >= 0 {
		tunesConf.Set("genesis_json_file", n.Genesisfile)
	}
	var pkey crypto.PrivKey = &edkey
	tunesConf.Set("gen_privkey", pkey)

	if err = agconf.InitRuntime(runtime, n.Chainid, tunesConf); err != nil {
		os.RemoveAll(n.ConfigPath)
	}
	return err
}

func DoInitNode(c *beego.Controller, runtime string) {
	n := &NodeInit{}
	c.ParseForm(n)
	if err := n.DoInit(runtime); err != nil {
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
	conf := viper.New()
	if len(n.ConfigPath) != 0 {
		conf.Set("runtime", n.ConfigPath)
	}
	chret := make(chan error, 0)
	go func() {
		if err := node.RunNodeRet(conf); err != nil {
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
	timer := time.NewTimer(time.Second)
	go func() {
		select {
		case <-timer.C:
			beego.BeeApp.Server.Close()
		}
	}()
	c.Data["json"] = "Server will be closed after 1 second,then you can close this page."
}
