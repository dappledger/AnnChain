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

package p2p

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"

	. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	cfg "github.com/dappledger/AnnChain/ann-module/lib/go-config"
	"github.com/dappledger/AnnChain/ann-module/lib/go-crypto"
)

type Reactor interface {
	Service // Start, Stop

	SetSwitch(*Switch)
	GetChannels() []*ChannelDescriptor
	AddPeer(peer *Peer)
	RemovePeer(peer *Peer, reason interface{})
	Receive(chID byte, peer *Peer, msgBytes []byte)
}

//--------------------------------------

type BaseReactor struct {
	BaseService // Provides Start, Stop, .Quit
	Switch      *Switch
}

func NewBaseReactor(logger *zap.Logger, name string, impl Reactor) *BaseReactor {
	return &BaseReactor{
		BaseService: *NewBaseService(logger, name, impl),
		Switch:      nil,
	}
}

func (br *BaseReactor) SetSwitch(sw *Switch) {
	br.Switch = sw
}
func (_ *BaseReactor) GetChannels() []*ChannelDescriptor              { return nil }
func (_ *BaseReactor) AddPeer(peer *Peer)                             {}
func (_ *BaseReactor) RemovePeer(peer *Peer, reason interface{})      {}
func (_ *BaseReactor) Receive(chID byte, peer *Peer, msgBytes []byte) {}

//-----------------------------------------------------------------------------

/*
The `Switch` handles peer connections and exposes an API to receive incoming messages
on `Reactors`.  Each `Reactor` is responsible for handling incoming messages of one
or more `Channels`.  So while sending outgoing messages is typically performed on the peer,
incoming messages are received on the reactor.
*/
type Switch struct {
	BaseService

	config       cfg.Config
	listeners    []Listener
	reactors     map[string]Reactor
	chDescs      []*ChannelDescriptor
	reactorsByCh map[byte]Reactor
	peers        *PeerSet
	dialing      *CMap
	nodeInfo     *NodeInfo             // our node info
	nodePrivKey  crypto.PrivKeyEd25519 // our node privkey

	filterConnByAddr       func(net.Addr) error
	filterConnByPubKey     func(crypto.PubKeyEd25519) error
	filterConnByRefuselist func(crypto.PubKeyEd25519) error

	authByCA       AuthorizationFunc
	authByCA_Local AuthorizationFunc

	addToRefuselist func([32]byte) error

	logger  *zap.Logger
	slogger *zap.SugaredLogger
}

var (
	ErrSwitchDuplicatePeer      = errors.New("Duplicate peer")
	ErrSwitchMaxPeersPerIPRange = errors.New("IP range has too many peers")
)

func NewSwitch(logger *zap.Logger, config cfg.Config) *Switch {
	setConfigDefaults(config)

	sw := &Switch{
		config:       config,
		reactors:     make(map[string]Reactor),
		chDescs:      make([]*ChannelDescriptor, 0),
		reactorsByCh: make(map[byte]Reactor),
		peers:        NewPeerSet(),
		dialing:      NewCMap(),
		nodeInfo:     nil,
		logger:       logger,
		slogger:      logger.Sugar(),
	}
	sw.BaseService = *NewBaseService(logger, "P2P Switch", sw)
	return sw
}

// Not goroutine safe.
func (sw *Switch) AddReactor(name string, reactor Reactor) Reactor {
	// Validate the reactor.
	// No two reactors can share the same channel.
	reactorChannels := reactor.GetChannels()
	for _, chDesc := range reactorChannels {
		chID := chDesc.ID
		if sw.reactorsByCh[chID] != nil {
			PanicSanity(fmt.Sprintf("Channel %X has multiple reactors %v & %v", chID, sw.reactorsByCh[chID], reactor))
		}
		sw.chDescs = append(sw.chDescs, chDesc)
		sw.reactorsByCh[chID] = reactor
	}
	sw.reactors[name] = reactor
	reactor.SetSwitch(sw)
	return reactor
}

// Not goroutine safe.
func (sw *Switch) Reactors() map[string]Reactor {
	return sw.reactors
}

// Not goroutine safe.
func (sw *Switch) Reactor(name string) Reactor {
	return sw.reactors[name]
}

// Not goroutine safe.
func (sw *Switch) AddListener(l Listener) {
	sw.listeners = append(sw.listeners, l)
}

// Not goroutine safe.
func (sw *Switch) Listeners() []Listener {
	return sw.listeners
}

// Not goroutine safe.
func (sw *Switch) IsListening() bool {
	return len(sw.listeners) > 0
}

// Not goroutine safe.
func (sw *Switch) SetNodeInfo(nodeInfo *NodeInfo) {
	sw.nodeInfo = nodeInfo
}

// Not goroutine safe.
func (sw *Switch) NodeInfo() *NodeInfo {
	return sw.nodeInfo
}

func (sw *Switch) NodePrivkey() crypto.PrivKeyEd25519 {
	return sw.nodePrivKey
}

func (sw *Switch) SetRefuseListFilter(cb func(crypto.PubKeyEd25519) error) {
	sw.filterConnByRefuselist = cb
}

// Not goroutine safe.
// NOTE: Overwrites sw.nodeInfo.PubKey
func (sw *Switch) SetNodePrivKey(nodePrivKey crypto.PrivKeyEd25519) {
	sw.nodePrivKey = nodePrivKey
	if sw.nodeInfo != nil {
		sw.nodeInfo.PubKey = nodePrivKey.PubKey().(crypto.PubKeyEd25519)
	}
}

// Switch.Start() starts all the reactors, peers, and listeners.
func (sw *Switch) OnStart() error {
	sw.BaseService.OnStart()
	// Start reactors
	for _, reactor := range sw.reactors {
		_, err := reactor.Start()
		if err != nil {
			return err
		}
	}
	// Start peers
	for _, peer := range sw.peers.List() {
		sw.startInitPeer(peer)
	}
	// Start listeners
	for _, listener := range sw.listeners {
		go sw.listenerRoutine(listener)
	}
	return nil
}

func (sw *Switch) OnStop() {
	sw.BaseService.OnStop()
	// Stop listeners
	for _, listener := range sw.listeners {
		listener.Stop()
	}
	sw.listeners = nil
	// Stop peers
	for _, peer := range sw.peers.List() {
		peer.Stop()
		sw.peers.Remove(peer)
	}
	// Stop reactors
	for _, reactor := range sw.reactors {
		reactor.Stop()
	}
}

// NOTE: This performs a blocking handshake before the peer is added.
// CONTRACT: Iff error is returned, peer is nil, and conn is immediately closed.
func (sw *Switch) AddPeerWithConnection(conn net.Conn, outbound bool) (*Peer, error) {

	// Filter by addr (ie. ip:port)
	if err := sw.FilterConnByAddr(conn.RemoteAddr()); err != nil {
		conn.Close()
		return nil, err
	}

	// Set deadline for handshake so we don't block forever on conn.ReadFull
	conn.SetDeadline(time.Now().Add(
		time.Duration(sw.config.GetInt(configKeyHandshakeTimeoutSeconds)) * time.Second))

	// First, encrypt the connection.
	var sconn net.Conn = conn
	if sw.config.GetBool(configKeyAuthEnc) {
		var err error
		sconn, err = MakeSecretConnection(conn, sw.nodePrivKey)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	if err := sw.FilterConnByRefuselist(sconn.(*SecretConnection).RemotePubKey()); err != nil {
		sconn.Close()
		return nil, err
	}

	// Filter by p2p-key
	if err := sw.FilterConnByPubKey(sconn.(*SecretConnection).RemotePubKey()); err != nil {
		sconn.Close()
		return nil, err
	}

	// Then, perform node handshake
	peerNodeInfo, err := peerHandshake(sconn, sw)
	if err != nil {
		sconn.Close()
		return nil, err
	}
	if sw.config.GetBool(configKeyAuthEnc) {
		// Check that the professed PubKey matches the sconn's.
		if !peerNodeInfo.PubKey.Equals(sconn.(*SecretConnection).RemotePubKey()) {
			sconn.Close()
			return nil, fmt.Errorf("Ignoring connection with unmatching pubkey: %v vs %v",
				peerNodeInfo.PubKey, sconn.(*SecretConnection).RemotePubKey())
		}
	}
	// Avoid self
	if peerNodeInfo.PubKey.Equals(sw.nodeInfo.PubKey) {
		sconn.Close()
		return nil, fmt.Errorf("Ignoring connection from self")
	}
	// Check version, chain id
	if err := sw.nodeInfo.CompatibleWith(peerNodeInfo); err != nil {
		sconn.Close()
		return nil, err
	}

	peer := newPeer(sw.logger, sw.config, sconn, peerNodeInfo, outbound, sw.reactorsByCh, sw.chDescs, sw.StopPeerForError)

	// Add the peer to .peers
	// ignore if duplicate or if we already have too many for that IP range
	if err := sw.peers.Add(peer); err != nil {
		sw.logger.Warn("Ignoring peer", zap.String("error", err.Error()), zap.Stringer("peer", peer))
		peer.Stop()
		return nil, err
	}

	// remove deadline and start peer
	conn.SetDeadline(time.Time{})
	if sw.IsRunning() {
		sw.startInitPeer(peer)
	}

	sw.logger.Info("Added peer", zap.Stringer("peer", peer))
	return peer, nil
}

func (sw *Switch) FilterConnByAddr(addr net.Addr) error {
	if sw.filterConnByAddr != nil {
		return sw.filterConnByAddr(addr)
	}
	return nil
}

func (sw *Switch) FilterConnByRefuselist(pubkey crypto.PubKeyEd25519) error {
	if sw.filterConnByRefuselist != nil {
		return sw.filterConnByRefuselist(pubkey)
	}
	return nil
}

func (sw *Switch) FilterConnByPubKey(pubkey crypto.PubKeyEd25519) error {
	if sw.filterConnByPubKey != nil {
		return sw.filterConnByPubKey(pubkey)
	}
	return nil

}
func (sw *Switch) AuthByCA_Local(locInfo *NodeInfo) error {
	if sw.authByCA_Local != nil {
		return sw.authByCA_Local(locInfo)
	}
	return nil
}

//func (sw *Switch) AuthByCA(peerInfo *NodeInfo) error {
//	if sw.authByCA != nil {
//		return sw.authByCA(peerInfo)
//	}
//	return nil
//}

func (sw *Switch) SetAuthByCA(f AuthorizationFunc) {
	sw.authByCA = f
}

func (sw *Switch) SetAuthByCA_Local(f AuthorizationFunc) {
	sw.authByCA_Local = f
}

func (sw *Switch) SetAddrFilter(f func(net.Addr) error) {
	sw.filterConnByAddr = f
}

func (sw *Switch) SetPubKeyFilter(f func(crypto.PubKeyEd25519) error) {
	sw.filterConnByPubKey = f
}

func (sw *Switch) SetAddToRefuselist(f func([32]byte) error) {
	sw.addToRefuselist = f
}

func (sw *Switch) AddToRefuselist(pk [32]byte) error {
	if sw.addToRefuselist != nil {
		return sw.addToRefuselist(pk)
	}
	return nil
}

func (sw *Switch) startInitPeer(peer *Peer) {
	peer.Start()               // spawn send/recv routines
	sw.addPeerToReactors(peer) // run AddPeer on each reactor
}

// Dial a list of seeds in random order
// Spawns a go routine for each dial
func (sw *Switch) DialSeeds(seeds []string) {
	// permute the list, dial them in random order.
	perm := rand.Perm(len(seeds))
	for i := 0; i < len(perm); i++ {
		go func(i int) {
			time.Sleep(time.Duration(rand.Int63n(3000)) * time.Millisecond)
			j := perm[i]
			addr := NewNetAddressString(seeds[j])

			sw.dialSeed(addr)
		}(i)
	}
}

func (sw *Switch) dialSeed(addr *NetAddress) {
	peer, err := sw.DialPeerWithAddress(addr)
	if err != nil {
		sw.logger.Error("Error dialing seed", zap.String("error", err.Error()))
		return
	} else {
		sw.logger.Info("Connected to seed", zap.Stringer("peer", peer))
	}
}

func (sw *Switch) DialPeerWithAddress(addr *NetAddress) (*Peer, error) {
	//sw.logger.Debug("Dialing address", zap.Stringer("address", addr))
	sw.dialing.Set(addr.IP.String(), addr)
	defer sw.dialing.Delete(addr.IP.String())

	conn, err := addr.DialTimeout(time.Duration(
		sw.config.GetInt(configKeyDialTimeoutSeconds)) * time.Second)
	if err != nil {
		sw.logger.Warn("Failed dialing address", zap.Stringer("address", addr), zap.String("error", err.Error()))
		return nil, err
	}
	if sw.config.GetBool(configFuzzEnable) {
		conn = FuzzConn(sw.config, conn)
	}
	peer, err := sw.AddPeerWithConnection(conn, true)
	if err != nil {
		sw.slogger.Warnw("Failed adding peer", "address", addr, "conn", conn, "error", err)
		return nil, err
	}
	sw.logger.Info("Dialed and added peer", zap.Stringer("address", addr), zap.Stringer("peer", peer))
	return peer, nil
}

func (sw *Switch) IsDialing(addr *NetAddress) bool {
	return sw.dialing.Has(addr.IP.String())
}

// Broadcast runs a go routine for each attempted send, which will block
// trying to send for defaultSendTimeoutSeconds. Returns a channel
// which receives success values for each attempted send (false if times out)
// NOTE: Broadcast uses goroutines, so order of broadcast may not be preserved.
func (sw *Switch) Broadcast(chID byte, msg interface{}) chan bool {
	successChan := make(chan bool, len(sw.peers.List()))
	//sw.slogger.Debugw("Broadcast", "channel", chID, "msg", msg)
	for _, peer := range sw.peers.List() {
		go func(peer *Peer) {
			success := peer.Send(chID, msg)
			successChan <- success
		}(peer)
	}
	return successChan
}

// Returns the count of outbound/inbound and outbound-dialing peers.
func (sw *Switch) NumPeers() (outbound, inbound, dialing int) {
	peers := sw.peers.List()
	for _, peer := range peers {
		if peer.outbound {
			outbound++
		} else {
			inbound++
		}
	}
	dialing = sw.dialing.Size()
	return
}

func (sw *Switch) Peers() IPeerSet {
	return sw.peers
}

// Disconnect from a peer due to external error.
// TODO: make record depending on reason.
func (sw *Switch) StopPeerForError(peer *Peer, reason interface{}) {
	sw.slogger.Infow("Stopping peer for error", "peer", peer, "error", reason)
	sw.peers.Remove(peer)
	peer.Stop()
	sw.removePeerFromReactors(peer, reason)
}

// Disconnect from a peer gracefully.
// TODO: handle graceful disconnects.
func (sw *Switch) StopPeerGracefully(peer *Peer) {
	sw.logger.Info("Stopping peer gracefully")
	sw.peers.Remove(peer)
	peer.Stop()
	sw.removePeerFromReactors(peer, nil)
}

func (sw *Switch) addPeerToReactors(peer *Peer) {
	for _, reactor := range sw.reactors {
		reactor.AddPeer(peer)
	}
}

func (sw *Switch) removePeerFromReactors(peer *Peer, reason interface{}) {
	for _, reactor := range sw.reactors {
		reactor.RemovePeer(peer, reason)
	}
}

func (sw *Switch) listenerRoutine(l Listener) {
	for {
		inConn, ok := <-l.Connections()
		if !ok {
			break
		}

		// ignore connection if we already have enough
		maxPeers := sw.config.GetInt(configKeyMaxNumPeers)
		if maxPeers <= sw.peers.Size() {
			sw.logger.Warn("Ignoring inbound connection: already have enough peers", zap.Stringer("address", inConn.RemoteAddr()), zap.Int("numPeers", sw.peers.Size()), zap.Int("max", maxPeers))
			continue
		}

		if sw.config.GetBool(configFuzzEnable) {
			inConn = FuzzConn(sw.config, inConn)
		}

		// New inbound connection!
		_, err := sw.AddPeerWithConnection(inConn, false)
		if err != nil {
			sw.logger.Warn("Ignoring inbound connection: error on AddPeerWithConnection", zap.Stringer("address", inConn.RemoteAddr()), zap.String("error", err.Error()))
			continue
		}

		// NOTE: We don't yet have the listening port of the
		// remote (if they have a listener at all).
		// The peerHandshake will handle that
	}

	// cleanup
}

//-----------------------------------------------------------------------------

type SwitchEventNewPeer struct {
	Peer *Peer
}

type SwitchEventDonePeer struct {
	Peer  *Peer
	Error interface{}
}

//------------------------------------------------------------------
// Switches connected via arbitrary net.Conn; useful for testing

// Returns n switches, connected according to the connect func.
// If connect==Connect2Switches, the switches will be fully connected.
// initSwitch defines how the ith switch should be initialized (ie. with what reactors).
// NOTE: panics if any switch fails to start.
func MakeConnectedSwitches(logger *zap.Logger, cfg *cfg.Config, n int, initSwitch func(int, *Switch) *Switch, connect func([]*Switch, int, int)) []*Switch {
	switches := make([]*Switch, n)
	for i := 0; i < n; i++ {
		switches[i] = makeSwitch(logger, cfg, i, "testing", "123.123.123", initSwitch)
	}

	if err := StartSwitches(switches); err != nil {
		panic(err)
	}

	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			connect(switches, i, j)
		}
	}

	return switches
}

var PanicOnAddPeerErr = false

// Will connect switches i and j via net.Pipe()
// Blocks until a conection is established.
// NOTE: caller ensures i and j are within bounds
func Connect2Switches(switches []*Switch, i, j int) {
	switchI := switches[i]
	switchJ := switches[j]
	c1, c2 := net.Pipe()
	doneCh := make(chan struct{})
	go func() {
		_, err := switchI.AddPeerWithConnection(c1, false) // AddPeer is blocking, requires handshake.
		if PanicOnAddPeerErr && err != nil {
			panic(err)
		}
		doneCh <- struct{}{}
	}()
	go func() {
		_, err := switchJ.AddPeerWithConnection(c2, true)
		if PanicOnAddPeerErr && err != nil {
			panic(err)
		}
		doneCh <- struct{}{}
	}()
	<-doneCh
	<-doneCh
}

func StartSwitches(switches []*Switch) error {
	for _, s := range switches {
		_, err := s.Start() // start switch and reactors
		if err != nil {
			return err
		}
	}
	return nil
}

func makeSwitch(logger *zap.Logger, cfg *cfg.Config, i int, network, version string, initSwitch func(int, *Switch) *Switch) *Switch {
	privKey := crypto.GenPrivKeyEd25519()
	// new switch, add reactors
	// TODO: let the config be passed in?
	s := initSwitch(i, NewSwitch(logger, *cfg))
	s.SetNodeInfo(&NodeInfo{
		PubKey:  privKey.PubKey().(crypto.PubKeyEd25519),
		Moniker: Fmt("switch%d", i),
		Network: network,
		Version: version,
	})
	s.SetNodePrivKey(privKey)
	return s
}
