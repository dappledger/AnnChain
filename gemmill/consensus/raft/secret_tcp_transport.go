package raft

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/p2p"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
)

type SecretTCPStreamLayer struct {
	privKey   crypto.PrivKey
	advertise net.Addr
	listener  *net.TCPListener
}

func (t *SecretTCPStreamLayer) Accept() (net.Conn, error) {
	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}
	return p2p.MakeSecretConnection(conn, t.privKey)
}

func (t *SecretTCPStreamLayer) Close() error {
	return t.listener.Close()
}

func (t *SecretTCPStreamLayer) Addr() net.Addr {

	if t.advertise != nil {
		return t.advertise
	}
	return t.listener.Addr()
}

func (t *SecretTCPStreamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (conn net.Conn, err error) {

	for i := 0; i < 3; i++ {

		conn, err = net.DialTimeout("tcp", string(address), timeout)
		if err != nil {
			log.Error("SecretTCPStreamLayer.Dial", zap.String("address", string(address)), zap.Error(err), zap.Int("try", i))
			time.Sleep(time.Second * 2)
			continue
		}
		return p2p.MakeSecretConnection(conn, t.privKey)
	}
	return
}

// NewTCPTransport returns a NetworkTransport that is built on top of
// a Secret TCP streaming transport layer.
func NewSecretTCPTransport(
	bindAddr string,
	advertise net.Addr,
	maxPool int,
	timeout time.Duration,
	logOutput io.Writer,
	privKey crypto.PrivKey,
) (*raft.NetworkTransport, error) {

	list, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}

	// Create stream
	stream := &SecretTCPStreamLayer{
		privKey:   privKey,
		advertise: advertise,
		listener:  list.(*net.TCPListener),
	}

	// Verify that we have a usable advertise address
	addr, ok := stream.Addr().(*net.TCPAddr)
	if !ok {
		list.Close()
		return nil, errors.New("not tcp")
	}
	if addr.IP.IsUnspecified() {
		list.Close()
		return nil, errors.New("errNotAdvertisable")
	}
	return raft.NewNetworkTransport(stream, maxPool, timeout, logOutput), nil
}
