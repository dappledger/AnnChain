package raft

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/hashicorp/raft"
)

type Peer struct {
	PubKey crypto.PubKey `json:"pub_key"`
	RPC    string        `json:"rpc"`
	Bind   string        `json:"bind"`
}

type ClusterConfig struct {
	filename string

	Local     Peer   `json:"local"`
	Advertise string `json:"advertise"`
	Peers     []Peer `json:"peers"`
}

func NewClusterConfig(filename string) (*ClusterConfig, error) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := ClusterConfig{}
	if err := wire.ReadJSONBytes(data, &c); err != nil {
		return nil, err
	}
	c.filename = filename
	return &c, nil
}

func (c *ClusterConfig) String() string {
	data := wire.JSONBytesPretty(c)
	return string(data)
}

func (c *ClusterConfig) Save() error {

	data := wire.JSONBytesPretty(c)

	tmp := c.filename + ".tmp"
	if err := ioutil.WriteFile(tmp, data, 0644); err != nil {
		return nil
	}

	if err := os.Rename(tmp, c.filename); err != nil {
		return err
	}
	return nil
}

func (c *ClusterConfig) AddPeer(peer Peer) bool {

	for _, p := range c.Peers {
		if p.Bind == peer.Bind {
			return false
		}
		if p.PubKey.Equals(peer.PubKey) {
			return false
		}
	}
	c.Peers = append(c.Peers, peer)
	return true
}

func (c *ClusterConfig) FindByBindAddress(bind string) *Peer {

	for i, p := range c.Peers {
		if p.Bind == bind {
			return &c.Peers[i]
		}

		n, err := tryResolveTCPAddr(p.Bind)
		if err != nil {
			continue
		}

		if bind == n {
			return &c.Peers[i]
		}
	}
	return nil
}

func (c *ClusterConfig) Remove(pubKey crypto.PubKey) bool {

	for i, p := range c.Peers {
		if p.PubKey.Equals(pubKey) {
			var peers []Peer
			if i > 0 {
				peers = c.Peers[:i]
			}
			if i != len(c.Peers)-1 {
				peers = append(peers, c.Peers[i+1:]...)
			}
			c.Peers = peers
			return true
		}
	}
	return false
}

func (c *ClusterConfig) LocalServer() raft.Server {

	return raft.Server{
		ID:      raft.ServerID(fmt.Sprintf("%x", c.Local.PubKey.Bytes())),
		Address: raft.ServerAddress(c.Local.Bind),
	}
}

func tryResolveTCPAddr(bind string) (string, error) {

	for i := 0; i < 10; i++ {
		n, err := net.ResolveTCPAddr("tcp", bind)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		return n.String(), nil
	}
	return "", fmt.Errorf("resolveTCPAddr %v err", bind)
}

func (c *ClusterConfig) Server() ([]raft.Server, error) {

	servers := make([]raft.Server, 0, len(c.Peers))
	for _, c := range c.Peers {

		n, err := tryResolveTCPAddr(c.Bind)
		if err != nil {
			return nil, fmt.Errorf("invalid address, err %v", err)
		}
		servers = append(servers, raft.Server{
			ID:      raft.ServerID(fmt.Sprintf("%x", c.PubKey.Bytes())),
			Address: raft.ServerAddress(n),
		})
	}
	return servers, nil
}
