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
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	cosiED "github.com/bford/golang-x-crypto/ed25519"
	"github.com/bford/golang-x-crypto/ed25519/cosi"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/ed25519"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
	"github.com/dappledger/AnnChain/module/xlib/def"
	"github.com/dappledger/AnnChain/src/tools"
	civiltypes "github.com/dappledger/AnnChain/src/types"
)

const (
	// CoSiTimeout sets the limits of the total timeout in seconds
	CoSiTimeout = 10

	// Packet types
	CoSiSyncMsg    = 0x01
	CoSiAckSyncMsg = 0x02
	RequestSigpart = 0x03

	_ byte = iota
	CommTypeIdentity
	CommTypeSecure
)

var (
	// CoSi related transaction tags
	CoSiTag     = []byte{'c', 'o', 's', 0x01}
	CoSiInitTag = []byte{'c', 'o', 's', 0x02}
)

type CoSiTx struct {
	civiltypes.CivilTx

	Type       string  `json:"type"` // what this cosi is for
	Leader     []byte  `json:"leader"`
	LeaderAddr string  `json:"leaderaddr"`
	TxHash     []byte  `json:"txhash"` // relevant tx
	Height     def.INT `json:"height"`
	Data       []byte  `json:"data"` // relevant data
}

type CoSiInitTx struct {
	civiltypes.CivilTx

	Type     string  `json:"type"`
	ChainID  string  `json:"chainid"`
	Receiver []byte  `json:"receiver"`
	TxHash   []byte  `json:"txhash"`
	Height   def.INT `json:"height"`
	Data     []byte  `json:"data"`
}

func IsCoSiTx(tx []byte) bool {
	return bytes.HasPrefix(tx, CoSiTag)
}

func IsCoSiInitTx(tx []byte) bool {
	return bytes.HasPrefix(tx, CoSiInitTag)
}

type CoSiModule struct {
	logger   *zap.Logger
	mtx      *sync.Mutex
	commuWG  *sync.WaitGroup
	privkey  cosiED.PrivateKey
	listener net.Listener
	valset   *agtypes.ValidatorSet

	cosig []byte

	CoSigners      *cosi.Cosigners
	Nonce          uint64
	Pubkeys        []cosiED.PublicKey
	Commitments    []cosi.Commitment
	SignatureParts []cosi.SignaturePart
	Conns          []*p2p.SecretConnection
}

type AggregateInfo struct {
	Pubkey []byte
	Commit []byte
}

type IdentityAck struct {
	Type      byte   `json:"type"`
	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`
}

func NewCoSiModule(logger *zap.Logger, privkey crypto.PrivKeyEd25519, validators *agtypes.ValidatorSet) *CoSiModule {
	numVal := validators.Size()
	mdl := &CoSiModule{
		mtx:      &sync.Mutex{},
		commuWG:  &sync.WaitGroup{},
		privkey:  privkey[:],
		listener: nil,
		logger:   logger,
		valset:   validators,

		Nonce:          rand.Uint64(),
		Conns:          make([]*p2p.SecretConnection, 0, numVal-1),
		Commitments:    make([]cosi.Commitment, 0, numVal),
		Pubkeys:        make([]cosiED.PublicKey, 0, numVal),
		SignatureParts: make([]cosi.SignaturePart, 0, numVal),
	}

	for _, v := range validators.Validators {
		pubkey := v.GetPubKey().(*crypto.PubKeyEd25519)
		mdl.Pubkeys = append(mdl.Pubkeys, pubkey[:])
	}

	mdl.CoSigners = cosi.NewCosigners(mdl.Pubkeys, nil)

	return mdl
}

func (cm *CoSiModule) SetLeaderListener(l net.Listener) {
	cm.listener = l
}

func (cm *CoSiModule) Lock() {
	cm.mtx.Lock()
}

func (cm *CoSiModule) Unlock() {
	cm.mtx.Unlock()
}

func (cm *CoSiModule) Add(n int) {
	cm.commuWG.Add(n)
}

func (cm *CoSiModule) Wait() {
	cm.commuWG.Wait()
}

func (cm *CoSiModule) Done() {
	cm.commuWG.Done()
}

func (cm *CoSiModule) Write(sc *p2p.SecretConnection, data []byte) (int, error) {
	nonceBytes := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(nonceBytes, cm.Nonce)
	nonceBytes = nonceBytes[:n]
	msg := make([]byte, binary.MaxVarintLen64, 1024)
	copy(msg[binary.MaxVarintLen64-n:binary.MaxVarintLen64], nonceBytes)
	msg = append(msg, data...)
	n, err := sc.Write(msg)
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (cm *CoSiModule) Read(sc *p2p.SecretConnection) ([]byte, error) {
	rBuf := make([]byte, 4096)
	rn, err := sc.Read(rBuf)
	if err != nil {
		return nil, err
	}
	if rn < binary.MaxVarintLen64 {
		return nil, fmt.Errorf("nonce needed")
	}

	s := 0
	for i, b := range rBuf[:binary.MaxVarintLen64] {
		if b != 0x00 {
			s = i
			break
		}
	}
	nonce, _ := binary.Uvarint(rBuf[s:binary.MaxVarintLen64])

	rData := rBuf[binary.MaxVarintLen64:rn]
	if len(rData) == 1 && rData[0] == CoSiSyncMsg {
		cm.Nonce = nonce
		cm.Write(sc, []byte{CoSiAckSyncMsg})
		return nil, nil
	}
	if nonce != cm.Nonce {
		return nil, fmt.Errorf("expect: %d, got: %d", cm.Nonce, nonce)
	}

	return rData, nil
}

func (cm *CoSiModule) identifyConnection(conn net.Conn) error {
	cryptoBytes := big.NewInt(rand.Int63()).Bytes()
	bufConn := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	bufConn.WriteByte(CommTypeIdentity)
	bufConn.Write(cryptoBytes)
	if err := bufConn.Flush(); err != nil {
		return err
	}
	rBuf := make([]byte, 1024)
	rn, err := bufConn.Read(rBuf)
	if err != nil && err != io.EOF {
		return err
	}
	id := &IdentityAck{}
	if err := json.Unmarshal(rBuf[:rn], id); err != nil {
		return err
	}
	pk, err := crypto.PubKeyFromBytes(id.Type, id.PubKey)
	if err != nil {
		return err
	}
	if !cm.valset.HasAddress(pk.Address()) {
		return fmt.Errorf("illegal connection, pubkey: %X is not a validator", id.PubKey)
	}
	pubkey := [32]byte(*(pk.(*crypto.PubKeyEd25519)))
	signature, err := crypto.SignatureFromBytes(id.Type, id.Signature)
	if err != nil {
		return err
	}
	sig := [64]byte(*(signature.(*crypto.SignatureEd25519)))
	if !ed25519.Verify(&pubkey, cryptoBytes, &sig) {
		return fmt.Errorf("illegal connection, signature failed")
	}

	return nil
}

func (cm *CoSiModule) secureConnection(conn net.Conn) (*p2p.SecretConnection, error) {
	privKey := crypto.PrivKeyEd25519{}
	copy(privKey[:], cm.privkey)
	sconn, err := p2p.MakeSecretConnection(conn, privKey)
	if err != nil {
		return nil, err
	}
	return sconn, nil
}

func (cm *CoSiModule) gatherCommitment(conn net.Conn, msg []byte) (net.Conn, error) {
	var err error
	if err = cm.identifyConnection(conn); err != nil {
		return conn, err
	}
	if _, err = conn.Write([]byte{CommTypeSecure}); err != nil {
		return conn, err
	}
	sconn, err := cm.secureConnection(conn)
	if err != nil {
		return conn, err
	}
	// if err = sconn.SetDeadline(time.Now().Add(CoSiTimeout * time.Second)); err != nil {
	// 	return sconn, err
	// }
	if _, err = cm.Write(sconn, []byte{CoSiSyncMsg}); err != nil {
		return sconn, err
	}
	res, _ := cm.Read(sconn)
	if len(res) != 1 || res[0] != CoSiAckSyncMsg {
		return sconn, errors.Errorf("missing sync ack msg")
	}
	requestData := append([]byte{RequestSigpart}, msg...)
	if _, err = cm.Write(sconn, requestData); err != nil {
		return sconn, err
	}
	res, err = cm.Read(sconn)
	if err != nil {
		return sconn, err
	}

	cm.Lock()
	cm.Conns = append(cm.Conns, sconn)
	cm.Commitments = append(cm.Commitments, cosi.Commitment(res))
	cm.Unlock()

	return sconn, nil
}

// LeadCoSign is called by the leader,
// which means leader automatically signs positive,
// otherwise the node should just refuse the CoSiTx without calling LeadCoSign
func (cm *CoSiModule) LeadCoSign(cosiAddr string, msg []byte) ([]byte, error) {
	protocol, address := tools.ProtocolAndAddress(cosiAddr)
	listener, _ := net.Listen(protocol, address)
	cm.SetLeaderListener(listener)

	numFollowers := int32(cm.valset.Size() - 1)
	finished := new(int32)

	leaderCommit, leaderSecret, _ := cosi.Commit(nil)
	cm.Commitments = append(cm.Commitments, leaderCommit)

	// soft
	ticker := time.NewTicker(500 * time.Millisecond)
	// hard limit
	due := time.AfterFunc(10*time.Second, func() {
		cm.logger.Debug("time is up, closing cosi listener")
		if err := cm.listener.Close(); err != nil {
			cm.logger.Error("timer fail to close cosi listener", zap.Error(err))
		}
	})

	go func() {
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(finished) == numFollowers {
					fmt.Println("ticker checks")
					ticker.Stop()
					if err := cm.listener.Close(); err != nil {
						cm.logger.Error("ticker fail to close cosi listener", zap.Error(err))
					}
					return
				}
			}
		}
	}()

	for {
		conn, err := cm.listener.Accept()
		if err != nil {
			break
		}

		go func() {
			if c, err := cm.gatherCommitment(conn, msg); err != nil {
				cm.logger.Error("cosi leader", zap.Error(err))
				c.Close()
				return
			}
			atomic.AddInt32(finished, 1)
		}()
	}

	if !due.Stop() {
		select {
		case <-due.C:
		default:
		}
	}

	// Compare the amount of commitments with the validators' quantity
	cm.Commitments = append(cm.Commitments)
	if len(cm.Commitments) < cm.valset.Size() {
		cm.logger.Error("lack of commitments")
		return nil, nil
	}

	aggr := AggregateInfo{
		Pubkey: []byte(cm.CoSigners.AggregatePublicKey()),
		Commit: cm.CoSigners.AggregateCommit(cm.Commitments),
	}
	aggrBytes, _ := json.Marshal(aggr)

	// leader's own signature part
	leaderSig := cosi.Cosign(cm.privkey[:], leaderSecret, msg, aggr.Pubkey, aggr.Commit)
	cm.SignatureParts = append(cm.SignatureParts, leaderSig)

	// Begin collecting round of cosi.SignaturePart.
	// Each follower contributes one signature part.
	cm.Add(int(numFollowers))
	for i := range cm.Conns {
		conn := cm.Conns[i]
		go func(sc *p2p.SecretConnection) {
			defer func() {
				// since secure connection has been established, any data can be trusted
				// so, it can be safely closed on any error
				sc.Close()
				cm.Done()
			}()

			if _, err := cm.Write(sc, aggrBytes); err != nil {
				cm.logger.Error("fail to write", zap.Error(err))
				return
			}
			spart, err := cm.Read(sc)
			if err != nil {
				cm.logger.Error("cosi error", zap.Error(err))
				return
			}
			cm.Lock()
			cm.SignatureParts = append(cm.SignatureParts, cosi.SignaturePart(spart))
			cm.Unlock()
		}(conn)
	}

	cm.Wait()

	cm.cosig = cm.CoSigners.AggregateSignature(aggr.Commit, cm.SignatureParts)
	var pks []cosiED.PublicKey
	for _, v := range cm.valset.Validators {
		pubkey := v.GetPubKey().(*crypto.PubKeyEd25519)
		pks = append(pks, pubkey[:])
	}

	if !cosi.Verify(pks, cosi.ThresholdPolicy(cm.valset.Size()*2/3+1), msg, cm.cosig) {
		fmt.Println("cosignature verification failed")
		return []byte{}, errors.Errorf("cosignature verification failed")
	}
	fmt.Println("cosi success")
	return cm.cosig, nil
}

// FollowCoSign is called by followers in a CoSi round
// Followers need to know the address of the leader and the actual data to be signed
// In FollowCoSign, followers will exchange their commitment with the leader and sign the data
func (cm *CoSiModule) FollowCoSign(leaderAddress string, data []byte) error {
	var conn net.Conn
	var err error
	retryLimit := 5
	retry := 0
	for {
		if retry == retryLimit {
			return fmt.Errorf("fail to setup the connection to the cosi leader: %v", err)
		}
		conn, err = net.Dial("tcp", leaderAddress)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
		retry++
	}

	rBuf := make([]byte, 1024)
	rn, err := conn.Read(rBuf)
	if err != nil {
		conn.Close()
		return errors.Wrap(err, "1")
	}
	if rBuf[0] != CommTypeIdentity {
		conn.Close()
		return fmt.Errorf("CommTypeIdentity first")
	}
	priv := [64]byte{}
	copy(priv[:], cm.privkey)
	privKey := crypto.PrivKeyEd25519(priv)
	sig := privKey.Sign(rBuf[1:rn])
	ack := IdentityAck{
		Type:      crypto.PubKeyTypeEd25519,
		PubKey:    privKey.PubKey().Bytes(),
		Signature: sig.Bytes(),
	}
	ackBytes, err := json.Marshal(ack)
	if err != nil {
		conn.Close()
		return errors.Wrap(err, "2")
	}
	if wn, err := conn.Write(ackBytes); err != nil {
		conn.Close()
		return errors.Wrap(err, "3")
	} else if len(ackBytes) != wn {
		conn.Close()
		return fmt.Errorf("incomplete write")
	}

	rn, err = conn.Read(rBuf)
	if rBuf[0] != CommTypeSecure {
		conn.Close()
		return fmt.Errorf("expected: CommTypeSecure, got: " + string(rBuf))
	}
	sc, err := cm.secureConnection(conn)
	if err != nil {
		conn.Close()
		return errors.Wrap(err, "4")
	}
	// if err = sc.SetDeadline(time.Now().Add(CoSiTimeout * time.Second)); err != nil {
	// 	sc.Close()
	// 	return err
	// }

	{
		// read CoSiSyncMsg and set up cosi nonce in the CoSiModule
		// result is nil, but it has to be the first call after secure connection is made
		_, err := cm.Read(sc)
		if err != nil {
			sc.Close()
			return errors.Wrap(err, "5")
		}
	}

	// now let's wait for the RequestSigpart
	res, err := cm.Read(sc)
	if err != nil {
		sc.Close()
		return errors.Wrap(err, "6")
	}

	if res[0] != RequestSigpart {
		sc.Close()
		return fmt.Errorf("request sigpart")
	}
	msg := res[1:]

	if !bytes.Equal(msg, data) {
		sc.Close()
		return fmt.Errorf("inconsistent txHash")
	}

	commit, secret, _ := cosi.Commit(nil)
	_, err = cm.Write(sc, []byte(commit))
	res, err = cm.Read(sc)
	if err != nil {
		sc.Close()
		return errors.Wrap(err, "7")
	}

	aggr := &AggregateInfo{}
	err = json.Unmarshal(res, aggr)
	if err != nil {
		sc.Close()
		return errors.Wrap(err, "8")
	}

	sigpart := cosi.Cosign(cm.privkey, secret, msg, cosiED.PublicKey(aggr.Pubkey), cosi.Commitment(aggr.Commit))
	_, err = cm.Write(sc, []byte(sigpart))

	if err = sc.Close(); err != nil {
		sc.Close()
		return errors.Wrap(err, "9")
	}

	return nil
}

// Verify the CoSi signature
func (cm *CoSiModule) Verify(sig []byte, msg []byte) bool {
	return cm.CoSigners.Verify(msg, sig)
}

const maxPacketLen = 512

var cosiHeader = []byte("|<*>|")

func IntToBytes(num int) []byte {
	intSize := unsafe.Sizeof(num)
	is := int(intSize)
	bs := make([]byte, is)

	for i := is - 1; i >= 0; i-- {
		t := num
		for j := 0; j < i; j++ {
			t = t >> 8
		}
		bs[i] = byte(t)
	}

	return bs
}

func BytesToInt(bs []byte) (ret int) {
	for i, b := range bs {
		it := int(b)
		for j := 0; j < i; j++ {
			it = it << 8
		}
		ret += it
	}

	return ret
}

type CosiPacket = []byte

func CosiWrapData(data []byte) []CosiPacket {
	intSize := int(0)
	intSize = int(unsafe.Sizeof(intSize))

	ret := make([]CosiPacket, 0, 1+len(data)/maxPacketLen)

	headPacket := append(cosiHeader, IntToBytes(len(data))...)
	headPacketLen := len(headPacket)

	cursor := maxPacketLen - headPacketLen
	if cursor > len(data) {
		return append(ret, append(headPacket, data...))
	}

	headPacket = append(headPacket, data[:cursor]...)
	ret = append(ret, headPacket)
	for {
		if cursor+512 > len(data) {
			ret = append(ret, data[cursor:])
			break
		} else {
			ret = append(ret, data[cursor:+512])
			cursor += 512
		}
	}

	return ret
}
