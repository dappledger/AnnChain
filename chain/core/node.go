// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
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

	"github.com/dappledger/AnnChain/chain/app"
	"github.com/dappledger/AnnChain/chain/types"
	"github.com/dappledger/AnnChain/gemmill"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/dappledger/AnnChain/gemmill/rpc/server"
	atypes "github.com/dappledger/AnnChain/gemmill/types"

	etypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/rlp"

	"github.com/spf13/viper"
)

const (
	ReceiptsPrefix = "receipts-"

	//RPCCollectSpecialVotes uint8 = iota
)

type Node struct {
	running       int64
	config        *viper.Viper
	privValidator *atypes.PrivValidator
	nodeInfo      *p2p.NodeInfo

	Angine      *gemmill.Angine
	AngineTune  *gemmill.Tunes
	Application atypes.Application
	GenesisDoc  *atypes.GenesisDoc
}

func queryPayLoadTxParser(txData []byte) ([]byte, error) {
	btx := etypes.Transaction{}
	err := rlp.DecodeBytes(txData, &btx)
	if err != nil {
		return nil, err
	}
	return btx.Data(), nil
}

func NewNode(conf *viper.Viper, runtime, appName string) (*Node, error) {

	// new app
	am, ok := app.AppMap[appName]
	if !ok {
		return nil, fmt.Errorf("App `%v` is not regiestered!", appName)
	}
	initApp, err := am(conf)
	if err != nil {
		return nil, fmt.Errorf("Create App instance error: %v", err)
	}

	// new angine
	tune := &gemmill.Tunes{Conf: conf, Runtime: runtime}
	//newAngine, err := gemmill.NewAngine(tune)
	newAngine, err := gemmill.NewAngine(initApp, tune)
	if err != nil {
		return nil, fmt.Errorf("new angine err", err)
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

	// newAngine.SetSpecialVoteRPC(node.GetSpecialVote)
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
		return fmt.Errorf("Failed to new node: %v", err)
	}
	if err := node.Start(); err != nil {
		return fmt.Errorf("Failed to start node: %v", err)
	}
	if config.GetString("rpc_laddr") != "" {
		if _, err := node.StartRPC(); err != nil {
			return fmt.Errorf("Failed to start rpc: %v", err)
		}
	}
	if config.GetBool("pprof") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	fmt.Printf("node (%s) is running on %s:%d ......\n", node.Angine.Genesis().ChainID, node.NodeInfo().ListenHost(), node.NodeInfo().ListenPort())

	gcmn.TrapSignal(func() {
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
			gcmn.Fmt("wire_version=%v", wire.Version),
			gcmn.Fmt("p2p_version=%v", p2p.Version),
			// gcmn.Fmt("consensus_version=%v", n.StateMachine.Version()),
			// gcmn.Fmt("rpc_version=%v/%v", rpc.Version, rpccore.Version),
			gcmn.Fmt("node_start_at=%s", strconv.FormatInt(time.Now().Unix(), 10)),
			gcmn.Fmt("commit_version=%s", types.GetCommitVersion()),
		},
		RemoteAddr: conf.GetString("rpc_laddr"),
		ListenAddr: gcmn.Fmt("%v:%v", p2pHost, p2pPort),
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
		rpcserver.RegisterRPCFuncs(mux, n.rpcRoutes())
		listener, err := rpcserver.StartHTTPServer(listenAddr, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}

	return listeners, nil
}

func (n *Node) PrivValidator() *atypes.PrivValidator {
	return n.privValidator
}

func (n *Node) HealthStatus() int {
	return n.Angine.HealthStatus()
}

func DefaultConf() *viper.Viper {
	globalConf := viper.New()
	// runtime, _ := cmd.Flags().GetString("runtime")

	globalConf.SetDefault("db_type", "sqlite3")
	globalConf.SetDefault("db_conn_str", "sqlite3") // some types of database will need this
	// globalConf.SetDefault("base_fee", 100)
	// globalConf.SetDefault("base_reserve", 10000000)
	// globalConf.SetDefault("max_txset_size", 10000)
	return globalConf
}
