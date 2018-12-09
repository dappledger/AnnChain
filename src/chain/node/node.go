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

package node

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine"
	ac "github.com/dappledger/AnnChain/angine/config"
	"github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
	//	client "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/module/lib/go-rpc/server"
	"github.com/dappledger/AnnChain/module/lib/go-wire"
	"github.com/dappledger/AnnChain/src/chain/log"
	"github.com/dappledger/AnnChain/src/chain/version"
)

const (
	ReceiptsPrefix          = "receipts-"
	OfficialAddress         = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
	ERR_NODE_ALREADY_RUNING = "node already running"
)

var node *Node

var Apps = make(map[string]AppMaker)

type Node struct {
	MainChainID string
	MainOrg     *OrgNode

	config        *viper.Viper
	privValidator *types.PrivValidator
	nodeInfo      *p2p.NodeInfo

	logger *zap.Logger
}

func AppExists(name string) (yes bool) {
	_, yes = Apps[name]
	return
}

func NewNode(conf *viper.Viper) *Node {
	aConf := ac.GetConfig(conf.GetString("runtime"))
	for k, v := range conf.AllSettings() {
		aConf.Set(k, v)
	}

	logger, err := NewLogger(aConf)
	if err != nil {
		fmt.Println("new logger err:", err)
		return nil
	}

	metropolis := NewMetropolis(logger, aConf)
	metroAngine := angine.NewAngine(logger, &angine.Tunes{Conf: aConf})
	tune := metroAngine.Tune
	if err := metroAngine.ConnectApp(metropolis); err != nil {
		cmn.PanicCrisis(err)
	}

	chainID := ""
	if metroAngine.Genesis() != nil {
		chainID = metroAngine.Genesis().ChainID
	}
	node := &Node{
		MainChainID: chainID,
		MainOrg: &OrgNode{
			Application: metropolis,
			Angine:      metroAngine,
			AngineTune:  tune,
			GenesisDoc:  metroAngine.Genesis(),
		},

		nodeInfo:      makeNodeInfo(aConf, metroAngine.PrivValidator().GetPubKey().(*crypto.PubKeyEd25519), metroAngine.P2PHost(), metroAngine.P2PPort()),
		config:        aConf,
		privValidator: metroAngine.PrivValidator(),
		logger:        logger,
	}

	// metroAngine.SetSpecialVoteRPC(node.GetSpecialVote)
	metroAngine.RegisterNodeInfo(node.nodeInfo)
	metropolis.SetNode(node)
	metropolis.SetCore(node.MainOrg)

	return node
}

func RunNode(config *viper.Viper) {
	if err := RunNodeRet(config); err != nil {
		cmn.Exit(cmn.Fmt("Failed to start node: %v", err))
	}
}

func NewLogger(conf *viper.Viper) (*zap.Logger, error) {
	env, logpath := conf.GetString("environment"), conf.GetString("log_path")
	if logpath == "" {
		var err error
		if logpath, err = os.Getwd(); err != nil {
			return nil, err
		}
	}
	viper.Set("log_path", logpath)
	return log.Initialize(env, path.Join(logpath, "node.output.log"), path.Join(logpath, "node.err.log")), nil
}

func RunNodeRet(config *viper.Viper) error {
	if node != nil {
		return errors.New(ERR_NODE_ALREADY_RUNING)
	}
	node = NewNode(config)
	if err := node.Start(); err != nil {
		return err
	}
	if node.GetConf().GetString("rpc_laddr") != "" {
		if _, err := node.StartRPC(); err != nil {
			return err
		}
	}
	if config.GetBool("pprof") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}
	fmt.Printf("node (%s) is running on %s:%d ......\n", node.MainChainID, node.NodeInfo().ListenHost(), node.NodeInfo().ListenPort())
	cmn.TrapSignal(func() {
		node.Stop()
	})
	return nil
}

// Call Start() after adding the listeners.
func (n *Node) Start() error {
	if err := n.MainOrg.Start(); err != nil {
		return fmt.Errorf("fail to start, error: %v", err)
	}

	n.MainOrg.GenesisDoc = n.MainOrg.Angine.Genesis()
	n.MainChainID = n.MainOrg.GenesisDoc.ChainID

	return nil
}

func (n *Node) Stop() {
	n.logger.Info("Stopping Node")
	n.MainOrg.Stop()
}

func makeNodeInfo(config *viper.Viper, pubkey *crypto.PubKeyEd25519, p2pHost string, p2pPort uint16) *p2p.NodeInfo {
	nodeInfo := &p2p.NodeInfo{
		PubKey:      *pubkey,
		Moniker:     config.GetString("moniker"),
		Network:     config.GetString("chain_id"),
		SigndPubKey: config.GetString("signbyCA"),
		Version:     version.GetVersion(),
		Other: []string{
			cmn.Fmt("wire_version=%v", wire.Version),
			cmn.Fmt("p2p_version=%v", p2p.Version),
			// Fmt("consensus_version=%v", n.StateMachine.Version()),
			// Fmt("rpc_version=%v/%v", rpc.Version, rpccore.Version),
			cmn.Fmt("node_start_at=%s", strconv.FormatInt(time.Now().Unix(), 10)),
			cmn.Fmt("revision=%s", version.GetCommitVersion()),
		},
		RemoteAddr: config.GetString("rpc_laddr"),
		ListenAddr: cmn.Fmt("%v:%v", p2pHost, p2pPort),
	}

	// We assume that the rpcListener has the same ExternalAddress.
	// This is probably true because both P2P and RPC listeners use UPnP,
	// except of course if the rpc is only bound to localhost

	return nodeInfo
}

func (n *Node) NodeInfo() *p2p.NodeInfo {
	return n.nodeInfo
}

func (n *Node) StartRPC() ([]net.Listener, error) {
	listenAddrs := strings.Split(n.config.GetString("rpc_laddr"), ",")
	listeners := make([]net.Listener, len(listenAddrs))

	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		// wm := rpcserver.NewWebsocketManager(rpcRoutes, n.evsw)
		// mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(n.logger, mux, n.rpcRoutes())
		listener, err := rpcserver.StartHTTPServer(n.logger, listenAddr, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}

	return listeners, nil
}

func (n *Node) PrivValidator() *types.PrivValidator {
	return n.privValidator
}

func (n *Node) GetConf() *viper.Viper {
	return n.config
}

// func (n *Node) GetSpecialVote(data []byte, validator *types.Validator) ([]byte, error) {
// 	clientJSON := client.NewClientJSONRPC(n.logger, validator.RPCAddress) // all shard nodes share the same rpc address of the Node
// 	tmResult := new(types.RPCResult)
// 	_, err := clientJSON.Call("vote_special_op", []interface{}{n.MainChainID, data}, tmResult)
// 	if err != nil {
// 		n.logger.Error("vote_special_op", zap.Error(err))
// 		return nil, err
// 	}
// 	res := (*tmResult).(*types.ResultRequestSpecialOP)
// 	if res.Code == types.CodeType_OK {
// 		return res.Data, nil
// 	}
// 	n.logger.Error("vote_special_op", zap.String("resultlog", res.Log))
// 	return nil, fmt.Errorf(res.Log)
// }

func CheckConfNeedInApp(appName string, conf map[string]interface{}) error {
	switch appName {
	case "evm":
		if _, ok := conf["cosi_laddr"]; !ok {
			return fmt.Errorf("cosi_laddr is missing,given available for gaining multisignature on event tx")
		}
		if _, ok := conf["event_laddr"]; !ok {
			return fmt.Errorf("event_laddr is missing,given available for dealing event tx")
		}
	case "ikhofi":
		if _, ok := conf["cosi_laddr"]; !ok {
			return fmt.Errorf("cosi_laddr is missing,given available for gaining multisignature on event tx")
		}
		if _, ok := conf["ikhofi_addr"]; !ok {
			return fmt.Errorf("ikhofi_addr is missing,given available for communicating with ikhofi server")
		}

	}
	return nil
}
