package raft

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/gemmill/blockchain"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/modules/go-common"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/state"
	"github.com/dappledger/AnnChain/gemmill/types"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/spf13/viper"
)

type config struct {
	raftDir            string // store logs, snapshot, peers
	peersDir           string
	snapshotDir        string
	snapshotLog        io.WriteCloser
	transLog           io.WriteCloser
	raftLog            io.WriteCloser
	blockSize          int
	blockPartSize      int
	emptyBlockInterval time.Duration
	clusterConfig      *ClusterConfig
	logStore           raft.LogStore
	stableStore        raft.StableStore
}

func initConfig(conf *viper.Viper) (*config, error) {

	runtime := conf.GetString("runtime")
	raftDir := path.Join(runtime, "raft")
	c := config{
		raftDir:     raftDir,
		peersDir:    runtime, // peers stored in $runtime/peers.json
		snapshotDir: path.Join(raftDir, "snapshot"),
	}

	raftLogDir := path.Join(raftDir, "logs")
	fs, err := os.Stat(raftLogDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(raftLogDir, os.ModeDir|0777); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if !fs.IsDir() {
		return nil, errors.New("raftDir is not empty and not dir")
	}

	{
		fname := path.Join(runtime, "raft", "logs", "snapshot.log")
		file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		c.snapshotLog = file

		logStore, err := raftboltdb.NewBoltStore(path.Join(runtime, "raft", "raft-log.bolt"))
		if err != nil {
			return nil, err
		}
		c.logStore = logStore

		stableStore, err := raftboltdb.NewBoltStore(path.Join(runtime, "raft", "raft-stable.bolt"))
		if err != nil {
			return nil, err
		}
		c.stableStore = stableStore
	}

	{
		fname := path.Join(runtime, "raft", "logs", "trans.log")
		file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		c.transLog = file
	}

	{
		fname := path.Join(raftLogDir, "raft.log")
		fp, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		c.raftLog = fp
	}

	clusterConfig, err := NewClusterConfig(path.Join(runtime, "raft-cluster.json"))
	if err != nil {
		return nil, err
	}
	c.clusterConfig = clusterConfig
	c.blockSize = conf.GetInt("block_size")
	c.blockPartSize = conf.GetInt("block_part_size")

	c.emptyBlockInterval = conf.GetDuration("raft.empty_block_interval")
	if c.emptyBlockInterval == 0 {
		c.emptyBlockInterval = time.Second * 3
	}

	return &c, nil
}

type ConsensusState struct {
	*common.BaseService
	rawRaft       *raft.Raft
	mempool       types.TxPool
	proposalCh    chan *types.Block
	evsw          types.EventSwitch
	privValidator *types.PrivValidator // for signing votes
	conf          *config
	stop          chan struct{}
	fsm           *BlockChainFSM
	mtx           sync.Mutex
	isRunning     uint32
}

func (cs *ConsensusState) OnStart() error {
	cs.BaseService.OnStart()

	return nil
}

func (cs *ConsensusState) OnStop() {
	cs.BaseService.OnStop()
	defer cs.conf.snapshotLog.Close()
	defer cs.conf.transLog.Close()
	cs.stop <- struct{}{}
}

func (cs *ConsensusState) GetValidators() (int64, []*types.Validator) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	return cs.fsm.state.LastBlockHeight, cs.fsm.state.Validators.Copy().Validators
}

func (cs *ConsensusState) SetEventSwitch(evsw types.EventSwitch) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.fsm.SetEventSwitch(evsw)
}

func NewConsensusState(vconf *viper.Viper, evsw types.EventSwitch, blockStore *blockchain.BlockStore, state *state.State, mempool types.TxPool, privValidator *types.PrivValidator) (*ConsensusState, error) {

	raftConf := raft.DefaultConfig()

	config, err := initConfig(vconf)
	if err != nil {
		return nil, err
	}
	raftConf.LogOutput = config.raftLog

	fsm := newBlockChainFSM(config, mempool, blockStore, state)

	raftConf.LocalID = config.clusterConfig.LocalServer().ID

	var a net.Addr
	if config.clusterConfig.Advertise != "" {
		a, err = net.ResolveTCPAddr("tcp", config.clusterConfig.Advertise)
		if err != nil {
			return nil, err
		}
	}

	trans, err := NewSecretTCPTransport(config.clusterConfig.Local.Bind, a, 10, time.Second*10, config.transLog, privValidator.PrivKey)
	if err != nil {
		return nil, err
	}

	fileSnap, err := raft.NewFileSnapshotStore(config.snapshotDir, 10, config.snapshotLog)
	if err != nil {
		return nil, err
	}

	// memStore is good enough, because Raft Log is same with the block, and block stored in LevelDB.
	// Besides, raft snapshot is useless,  BlockChain sync system did it for snapshot.
	//store := raft.NewInmemStore()

	rawRaft, err := raft.NewRaft(raftConf, fsm, config.logStore, config.stableStore, fileSnap, trans)
	if err != nil {
		return nil, err
	}
	server, err := config.clusterConfig.Server()
	if err != nil {
		return nil, err
	}
	configuration := raft.Configuration{Servers: server}
	fmt.Println("server:", server)

	rawRaft.BootstrapCluster(configuration)

	cs := &ConsensusState{
		rawRaft:       rawRaft,
		mempool:       mempool,
		proposalCh:    make(chan *types.Block),
		evsw:          evsw,
		conf:          config,
		stop:          make(chan struct{}),
		fsm:           fsm,
		privValidator: privValidator,
	}
	cs.BaseService = common.NewBaseService("RaftConsensusState", cs)

	types.AddListenerForEvent(cs.evsw, "conS", types.EventStringSwitchToConsensus(), func(data types.TMEventData) {
		go cs.run()
	})

	return cs, err
}

func (cs *ConsensusState) getSignature(b *types.Block) (crypto.Signature, error) {

	return crypto.SignatureFromBytes(b.Extra)
}

func (cs *ConsensusState) sign(b *types.Block) {

	b.Header.Extra = cs.privValidator.Sign(b.Hash()).Bytes()
}

func (cs *ConsensusState) ValidateBlock(b *types.Block) error {

	s := cs.fsm.state

	if err := b.ValidateBasic(s.ChainID, s.LastBlockHeight, s.LastBlockID, s.LastBlockTime, s.AppHash, s.ReceiptsHash); err != nil {
		return err
	}

	sig, err := cs.getSignature(b)
	if err != nil {
		return err
	}

	_, v := cs.fsm.state.Validators.GetByAddress(b.ProposerAddress)
	if !v.PubKey.VerifyBytes(b.Hash(), sig) {
		return errors.New(common.Fmt("Wrong Block.Signature.  proposerAddress %x, hash %x, signature %v", b.ProposerAddress, b.Hash(), sig.String()))
	}

	return nil
}

func (cs *ConsensusState) run() {

	if !atomic.CompareAndSwapUint32(&cs.isRunning, 0, 1) {
		return
	}

L1:
	for {
		select {
		case _ = <-cs.stop:
			break L1
		default:
			switch cs.rawRaft.State() {
			case raft.Follower:
				select {
				case _ = <-cs.fsm.AppliedCh():
				case _ = <-time.After(time.Second * 1):
				}
			case raft.Leader:

				start := time.Now()
				b := cs.fsm.createProposalBlock(cs.privValidator.GetAddress())
				end := time.Now()
				log.Debug("createProposalBlock", zap.Int64("height", b.Height), zap.Duration("taken", end.Sub(start)))
				cs.sign(b)

				data := wire.BinaryBytes(b)
				cs.rawRaft.Apply(data, time.Second*3)

				_ = <-cs.fsm.AppliedCh()
			default:
				time.Sleep(time.Second)
			}
		}
	}
	atomic.StoreUint32(&cs.isRunning, 0)
}

func (cs *ConsensusState) NewPublicAPI() *PublicAPI {
	return &PublicAPI{cs}
}
