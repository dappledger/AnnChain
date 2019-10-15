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

package core

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"

	"github.com/dappledger/AnnChain/chain/app"
	"github.com/dappledger/AnnChain/chain/types"
	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/core/vm"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	cmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/rpc/server"
	gtypes "github.com/dappledger/AnnChain/gemmill/types"
)

type Node struct {
	running       int64
	config        *viper.Viper
	privValidator *gtypes.PrivValidator
	nodeInfo      *p2p.NodeInfo

	Angine      *gemmill.Angine
	AngineTune  *gemmill.Tunes
	Application gtypes.Application
	GenesisDoc  *gtypes.GenesisDoc
}

func queryPayLoadTxParser(txData []byte) ([]byte, error) {
	btx := etypes.Transaction{}
	err := rlp.DecodeBytes(txData, &btx)
	if err != nil {
		return nil, err
	}
	return btx.Data(), nil
}

func (nd *Node) ExecAdminTx(app *vm.AdminDBApp, tx []byte) error {
	return nd.Angine.ExecAdminTx(app, tx)
}

func NewNode(conf *viper.Viper, runtime, appName string) (*Node, error) {

	// new app
	am, ok := app.AppMap[appName]
	if !ok {
		return nil, fmt.Errorf("app `%v` is not regiestered", appName)
	}
	initApp, err := am(conf)
	if err != nil {
		return nil, fmt.Errorf("create App instance error: %v", err)
	}

	// new angine
	tune := &gemmill.Tunes{Conf: conf, Runtime: runtime}
	newAngine, err := gemmill.NewAngine(initApp, tune)
	if err != nil {
		return nil, fmt.Errorf("new angine error: %v", err)
	}
	newAngine.SetQueryPayLoadTxParser(queryPayLoadTxParser)

	newAngine.ConnectApp(initApp)

	node := &Node{
		Application: initApp,
		Angine:      newAngine,
		AngineTune:  tune,
		GenesisDoc:  newAngine.Genesis(),

		nodeInfo:      makeNodeInfo(conf, newAngine.PrivValidator().PubKey, newAngine.P2PHost(), newAngine.P2PPort()),
		config:        conf,
		privValidator: newAngine.PrivValidator(),
	}
	vm.DefaultAdminContract.SetCallback(node.ExecAdminTx)
	// newAngine.SetAdminVoteRPC(node.GetAdminVote)
	newAngine.RegisterNodeInfo(node.nodeInfo)
	initApp.SetCore(newAngine)

	return node, nil
}

func RunNode(config *viper.Viper, runtime, appName string) {
	if err := RunNodeRet(config, runtime, appName); err != nil {
		panic(err)
	}
}

func RunNodeRet(config *viper.Viper, runtime, appName string) error {
	node, err := NewNode(config, runtime, appName)
	if err != nil {
		return fmt.Errorf("failed to new node: %v", err)
	}
	if err := node.Start(); err != nil {
		return fmt.Errorf("failed to start node: %v", err)
	}
	if config.GetString("rpc_laddr") != "" {
		if _, err := node.StartRPC(); err != nil {
			return fmt.Errorf("failed to start rpc: %v", err)
		}
	}
	if config.GetBool("pprof") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	fmt.Printf("node (%s) is running on %s:%d ......\n", node.Angine.Genesis().ChainID, node.NodeInfo().ListenHost(), node.NodeInfo().ListenPort())

	cmn.TrapSignal(func() {
		node.Stop()
	})
	return nil
}

// Call Start() after adding the listeners.
func (n *Node) Start() error {
	if !atomic.CompareAndSwapInt64(&n.running, 0, 1) {
		return fmt.Errorf("already started")
	}

	n.Application.Start()
	if err := n.Angine.Start(); err != nil {
		return fmt.Errorf("fail to start, error: %v", err)
	}

	// TODO timeout
	for n.Angine.NoneGenesis() {
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (n *Node) Stop() {
	log.Info("Stopping Node")

	if atomic.CompareAndSwapInt64(&n.running, 1, 0) {
		n.Application.Stop()
		n.Angine.Stop()
	}
}

func (n *Node) IsRunning() bool {
	return atomic.LoadInt64(&n.running) == 1
}

func makeNodeInfo(conf *viper.Viper, pubkey crypto.PubKey, p2pHost string, p2pPort uint16) *p2p.NodeInfo {
	nodeInfo := &p2p.NodeInfo{
		PubKey:      pubkey,
		Moniker:     conf.GetString("moniker"),
		Network:     conf.GetString("chain_id"),
		SigndPubKey: conf.GetString("signbyCA"),
		Version:     types.GetVersion(),
		Other: []string{
			cmn.Fmt("wire_version=%v", wire.Version),
			cmn.Fmt("p2p_version=%v", p2p.Version),
			// cmn.Fmt("consensus_version=%v", n.StateMachine.Version()),
			// cmn.Fmt("rpc_version=%v/%v", rpc.Version, rpccore.Version),
			cmn.Fmt("node_start_at=%s", strconv.FormatInt(time.Now().Unix(), 10)),
			cmn.Fmt("commit_version=%s", types.GetCommitVersion()),
		},
		RemoteAddr: conf.GetString("rpc_laddr"),
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
		routes := n.rpcRoutes()
		for _, v := range n.Angine.APIs() {
			for n, h := range v {
				routes[n] = h
			}
		}
		rpcserver.RegisterRPCFuncs(mux, routes)

		listener, err := rpcserver.StartHTTPServer(listenAddr, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}

	return listeners, nil
}

func (n *Node) PrivValidator() *gtypes.PrivValidator {
	return n.privValidator
}

func (n *Node) HealthStatus() int {
	return n.Angine.HealthStatus()
}

//func (n *Node) GetAdminVote(data []byte, validator *gtypes.Validator) ([]byte, error) {
//	clientJSON := client.NewClientJSONRPC(validator.RPCAddress) // all shard nodes share the same rpc address of the Node
//	rpcResult := new(gtypes.RPCResult)
//	_, err := clientJSON.Call("vote_admin_op", []interface{}{n.GenesisDoc.ChainID, data}, rpcResult)
//	if err != nil {
//		return nil, err
//	}
//	res := (*rpcResult).(*gtypes.ResultRequestAdminOP)
//	if res.Code == gtypes.CodeType_OK {
//		return res.Data, nil
//	}
//	return nil, fmt.Errorf(res.Log)
//}

func DefaultConf() *viper.Viper {
	globalConf := viper.New()
	// runtime, _ := cmd.Flags().GetString("runtime")

	return globalConf
}
