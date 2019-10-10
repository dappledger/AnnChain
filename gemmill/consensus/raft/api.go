package raft

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	rpcclient "github.com/dappledger/AnnChain/gemmill/rpc/client"
	rpcserver "github.com/dappledger/AnnChain/gemmill/rpc/server"
	"github.com/dappledger/AnnChain/gemmill/types"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type PublicAPI struct {
	*ConsensusState
}

func (p *PublicAPI) API() map[string]*rpcserver.RPCFunc {

	return map[string]*rpcserver.RPCFunc{
		"raft/role":        rpcserver.NewRPCFunc(p.Role, ""),
		"raft/leader":      rpcserver.NewRPCFunc(p.Leader, ""),
		"raft/add_peer":    rpcserver.NewRPCFunc(p.AddPeer, "addr,rpc,pubKey"),
		"raft/remove_peer": rpcserver.NewRPCFunc(p.RemovePeer, "pubKey"),
		"raft/stats":       rpcserver.NewRPCFunc(p.Stats, ""),
	}
}

func (p *PublicAPI) Role() (string, error) {

	return p.rawRaft.State().String(), nil
}

type LeaderResult struct {
	ID      string
	Address string
	RPC     string
}

func (p *PublicAPI) Leader() (*LeaderResult, error) {

	if p.rawRaft.Leader() == "" {
		return nil, errors.New("not found leader")

	}

	peer := p.conf.clusterConfig.FindByBindAddress(string(p.rawRaft.Leader()))
	if peer == nil {
		return nil, errors.New("not found leader")
	}

	rpc, err := tryResolveTCPAddr(peer.RPC)
	if err != nil {
		return nil, err
	}
	bind, err := tryResolveTCPAddr(peer.Bind)
	if err != nil {
		return nil, err
	}
	return &LeaderResult{ID: fmt.Sprintf("%x", peer.PubKey.Bytes()), Address: string(bind), RPC: rpc}, nil
}

type AddPeerResult struct{}

func (p *PublicAPI) AddPeer(addr, rpc string, pubKey string) (*AddPeerResult, error) {

	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, fmt.Errorf("invalid pubKey, must be hex encode")
	}

	if p.rawRaft.State() == raft.Leader {

		for _, peer := range p.conf.clusterConfig.Peers {

			if p.conf.clusterConfig.Local.PubKey.Equals(peer.PubKey) {
				continue
			}

			cli := rpcclient.NewClientJSONRPC(peer.RPC)
			r := AddPeerResult{}
			if _, err := cli.Call("raft/add_peer", []interface{}{addr, rpc, pubKey}, &r); err != nil {
				log.Error("call raft add peer", zap.String("rpc", peer.RPC), zap.Error(err))
			}
		}
		i := p.rawRaft.AddVoter(raft.ServerID(pubKey), raft.ServerAddress(addr), 0, 0)
		if err := i.Error(); err != nil {
			return nil, err
		}
	}

	publicKey, err := crypto.PubKeyFromBytes(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	if p.conf.clusterConfig.AddPeer(Peer{PubKey: publicKey, RPC: rpc, Bind: addr}) {
		if err := p.conf.clusterConfig.Save(); err != nil {
			return nil, err
		}
	}

	p.ConsensusState.fsm.state.Validators.Add(&types.Validator{
		Address: publicKey.Address(),
		PubKey:  publicKey,
	})
	return &AddPeerResult{}, nil
}

func (p *PublicAPI) Stats() (string, error) {
	bs, err := json.MarshalIndent(p.rawRaft.Stats(), "", "\t")
	return string(bs), err
}

type RemovePeerResult struct{}

func (p *PublicAPI) RemovePeer(pubKey string) (*RemovePeerResult, error) {

	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, fmt.Errorf("invalid pubKey, must be hex encode")
	}

	if p.rawRaft.State() == raft.Leader {

		for _, peer := range p.conf.clusterConfig.Peers {

			if p.conf.clusterConfig.Local.PubKey.Equals(peer.PubKey) {
				continue
			}

			cli := rpcclient.NewClientJSONRPC(peer.RPC)
			r := RemovePeerResult{}
			if _, err := cli.Call("raft/remove_peer", []interface{}{pubKey}, &r); err != nil {
				log.Error("call peers remove peer", zap.String("rpc", peer.RPC), zap.Error(err))
			}
		}

		i := p.rawRaft.RemoveServer(raft.ServerID(pubKey), 0, 0)
		if err := i.Error(); err != nil {
			return nil, err
		}
	}

	publicKey, err := crypto.PubKeyFromBytes(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	if p.conf.clusterConfig.Remove(publicKey) {
		if err := p.conf.clusterConfig.Save(); err != nil {
			return nil, err
		}
	}

	// We can not remove it from validators, otherwise when a new node replay txs, it will be got errors
	// p.ConsensusState.fsm.state.Validators.Remove(publicKey.Address())
	return &RemovePeerResult{}, nil
}
