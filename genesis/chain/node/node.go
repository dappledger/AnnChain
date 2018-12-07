package node

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine"
	at "github.com/dappledger/AnnChain/angine/types"
	cmn "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/ann-module/lib/go-crypto"
	"github.com/dappledger/AnnChain/ann-module/lib/go-p2p"
	client "github.com/dappledger/AnnChain/ann-module/lib/go-rpc/client"
	"github.com/dappledger/AnnChain/ann-module/lib/go-rpc/server"
	"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
	"github.com/dappledger/AnnChain/genesis/chain/version"
)

type Node struct {
	running       int64
	config        cfg.Config
	privValidator *at.PrivValidator
	nodeInfo      *p2p.NodeInfo
	logger        *zap.Logger
	Angine        *angine.Angine
	AngineTune    *angine.AngineTunes
	Application   at.Application
	GenesisDoc    *at.GenesisDoc
}

func NewNode(logger *zap.Logger, config cfg.Config, initApp at.Application) *Node {
	conf := config.(*cfg.MapConfig)
	tune := &angine.AngineTunes{Conf: conf}
	newAngine := angine.NewAngine(tune)
	newAngine.ConnectApp(initApp)

	node := &Node{
		Application: initApp,
		Angine:      newAngine,
		AngineTune:  tune,
		GenesisDoc:  newAngine.Genesis(),

		nodeInfo:      makeNodeInfo(conf, newAngine.PrivValidator().PubKey.(crypto.PubKeyEd25519), newAngine.P2PHost(), newAngine.P2PPort()),
		config:        conf,
		privValidator: newAngine.PrivValidator(),
		logger:        logger,
	}

	newAngine.SetSpecialVoteRPC(node.GetSpecialVote)
	newAngine.RegisterNodeInfo(node.nodeInfo)

	return node
}

func RunNode(logger *zap.Logger, config cfg.Config, initApp at.Application) {
	node := NewNode(logger, config, initApp)
	if err := node.Start(); err != nil {
		cmn.Exit(cmn.Fmt("Failed to start node: %v", err))
	}
	if config.GetString("rpc_laddr") != "" {
		if _, err := node.StartRPC(); err != nil {
			cmn.PanicCrisis(err)
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

	return nil
}

func (n *Node) Stop() {
	n.logger.Info("Stopping Node")

	if atomic.CompareAndSwapInt64(&n.running, 1, 0) {
		n.Application.Stop()
		n.Angine.Stop()
	}
}

func (n *Node) IsRunning() bool {
	return atomic.LoadInt64(&n.running) == 1
}

func makeNodeInfo(config cfg.Config, pubkey crypto.PubKeyEd25519, p2pHost string, p2pPort uint16) *p2p.NodeInfo {
	nodeInfo := &p2p.NodeInfo{
		PubKey:      pubkey,
		Moniker:     config.GetString("moniker"),
		Network:     config.GetString("chain_id"),
		SigndPubKey: config.GetString("signbyCA"),
		Version:     version.GetVersion(),
		Other: []string{
			cmn.Fmt("wire_version=%v", wire.Version),
			cmn.Fmt("p2p_version=%v", p2p.Version),
			cmn.Fmt("node_start_at=%s", strconv.FormatInt(time.Now().Unix(), 10)),
			cmn.Fmt("revision=%s", version.GetCommitVersion()),
		},
		RemoteAddr: config.GetString("rpc_laddr"),
		ListenAddr: cmn.Fmt("%v:%v", p2pHost, p2pPort),
	}

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
		rpcserver.RegisterRPCFuncs(n.logger, mux, n.rpcRoutes())
		listener, err := rpcserver.StartHTTPServer(n.logger, listenAddr, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}

	return listeners, nil
}

func (n *Node) PrivValidator() *at.PrivValidator {
	return n.privValidator
}

func (n *Node) GetSpecialVote(data []byte, validator *at.Validator) ([]byte, error) {
	clientJSON := client.NewClientJSONRPC(n.logger, validator.RPCAddress) // all shard nodes share the same rpc address of the Node
	tmResult := new(at.RPCResult)
	_, err := clientJSON.Call("vote_special_op", []interface{}{data}, tmResult)
	if err != nil {
		return nil, err
	}
	res := (*tmResult).(*at.ResultRequestSpecialOP)
	if res.Code == at.CodeType_OK {
		return res.Data, nil
	}
	return nil, fmt.Errorf(res.Log)
}
