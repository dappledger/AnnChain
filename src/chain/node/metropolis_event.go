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
	"bytes"
	//	"encoding/gob"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"

	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-p2p"
	"github.com/dappledger/AnnChain/src/tools"
	cvtools "github.com/dappledger/AnnChain/src/tools"
	cvtypes "github.com/dappledger/AnnChain/src/types"
)

func (met *Metropolis) CodeExists(codehash []byte) bool {
	code := met.EventCodeBase.Get(codehash)
	return !(code == nil || len(code) == 0)
}

// PublishEvent implements node.Publisher interface.
// all Apps running on this node call this Publish to publish event if they want
// now we just publish the event with no difference, maybe we could offer some customized subscription
func (met *Metropolis) PublishEvent(chainID string, block *agtypes.BlockCache, data []EventData, relatedTxHash []byte) error {
	accnt, err := met.EventState.GetAccount(chainID)
	if err != nil {
		met.logger.Warn("PublishEvent error", zap.Error(err), zap.String("chainid", chainID))
		return nil
	}
	subscribers := accnt.GetSubscribers()
	if len(subscribers) == 0 {
		met.logger.Warn("PublishEvent: no subscribers")
		return nil
	}

	privkey, _ := met.core.GetPrivateKey()
	pubkey, _ := met.core.GetPublicKey()

	bheader := block.Header
	sendNotification := bytes.Equal(pubkey[:], bheader.Maker)

	for _, dat := range data {
		batch := met.EventWarehouse.Batch()
		notifications := make([]*EventNotificationTx, 0)
		for subChainID, codeHash := range subscribers {
			var resStack cvtypes.ParamUData
			code := met.EventCodeBase.Get(codeHash)

			if len(code) > 0 {
				luas := cvtools.NewLuaState()
				var err error
				if resStack, err = cvtools.ExecLuaWithParam(luas, string(code), dat); err != nil {
					met.logger.Warn("[publish_event],exec lua code err", zap.Error(err))
					continue
				}
			}

			if len(resStack) == 0 {
				fmt.Println("this subscriber doesn't care about this")
				continue // this subscriber doesn't care about this
			}

			eventData := make(EventData)
			for key, val := range resStack {
				if lud, ok := val.(*lua.LUserData); ok {
					eventData[key] = lud.Value
				} else {
					eventData[key] = val
				}
			}

			datBytes, err := EncodeEventData(eventData)
			if err != nil {
				met.logger.Warn("[publish_event],encodeEventData failed", zap.Error(err))
				continue
			}
			datahash, _ := tools.HashKeccak(datBytes)

			met.EventWarehouse.Push(batch, subChainID, chainID, bheader.Height, datBytes)

			if sendNotification {
				ntx := &EventNotificationTx{
					Listener:  subChainID,
					Source:    chainID,
					Height:    bheader.Height,
					DataHash:  datahash,
					RelatedTx: relatedTxHash,
					Time:      time.Now(),
				}
				notifications = append(notifications, ntx)
			}
		}
		met.EventWarehouse.Flush(batch)

		if sendNotification && len(notifications) > 0 {
			for _, n := range notifications {
				n.Signature, _ = cvtools.TxSign(n, &privkey)
				txBytes, _ := cvtools.TxToBytes(n)
				if err := met.BroadcastTx(agtypes.WrapTx(EventNotificationTag, txBytes)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (met *Metropolis) spawnOffchainEventListener() error {
	if !met.config.IsSet("event_laddr") {
		return errors.Errorf(`config "event_laddr" is missing`)
	}
	eventExternal := met.config.GetString("event_laddr")

	protocol, address := tools.ProtocolAndAddress(eventExternal)
	lAddrIP, lAddrPort, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	listener, err := net.Listen(protocol, address)
	if err != nil {
		return err
	}

	eventExternal, err = tools.DetermineExternalAddress(listener, lAddrIP, lAddrPort, met.config.GetBool("skip_upnp"))
	if err != nil {
		listener.Close()
		return errors.Wrap(err, "fail to determine the event external address")
	}
	met.config.Set("event_external_address", eventExternal)

	go func() {
		for {
			conn, err := listener.Accept()
			if err == nil {
				go met.handleEventConnection(conn)
			} else {
				met.logger.Error("error on accept new conn from event listner", zap.Error(err))
			}
		}
	}()

	return nil
}

func (met *Metropolis) handleEventConnection(conn net.Conn) {
	privkey, _ := met.core.GetPrivateKey()
	// pubkey, _ := met.core.GetPublicKey()
	sc, err := p2p.MakeSecretConnection(conn, privkey)
	if err != nil {
		met.logger.Error("fail to make secret connection", zap.Error(err))
		conn.Close()
		return
	}
	var indexBuf []byte
	{
		rBuf := make([]byte, 4096)
		rn, err := sc.Read(rBuf)
		if err != nil {
			met.logger.Error("fail to read event index", zap.Error(err))
			sc.Close()
			return
		}
		indexBuf = rBuf[:rn]
	}
	strs := strings.Split(string(indexBuf), ",")
	if len(strs) != 3 {
		met.logger.Debug("illegal event index", zap.Strings("index", strs))
		sc.Close()
		return
	}
	subID, pubID, heightStr := strs[0], strs[1], strs[2]
	remotePubKey := sc.RemotePubKey()
	if !met.OrgState.ExistAccount(subID) {
		met.logger.Debug("missing subID in orgState")
		sc.Close()
		return
	}
	remoteAccnt, _ := met.OrgState.GetAccount(subID)
	remoteNodes := remoteAccnt.GetNodes()
	if _, ok := remoteNodes[remotePubKey.KeyString()]; !ok {
		met.logger.Warn("illegal node (" + remotePubKey.KeyString() + ") trying to fetch event data")
		sc.Close()
		return
	}
	height, err := agtypes.AtoHeight(heightStr)
	if err != nil {
		met.logger.Debug("height is not a number: ", zap.Error(err))
		sc.Close()
		return
	}
	data, err := met.EventWarehouse.Fetch(subID, pubID, height)
	if len(data) == 0 || err != nil {
		met.logger.Error("fail to fetch the event data", zap.Error(err))
		sc.Close()
		return
	}

	if _, err = sc.Write(data); err != nil {
		met.logger.Error("fail to send event data", zap.Error(err))
		sc.Close()
		return
	}

	var emsg []byte
	{
		mBuf := make([]byte, 1024)
		msg, err := sc.Read(mBuf)
		if err != nil {
			met.logger.Error("fail to read msg", zap.Error(err))
			sc.Close()
			return
		}

		emsg = mBuf[:msg]
		smsg := string(emsg)
		if smsg != "event has been confirm" {
			met.logger.Error("msg is wrong", zap.Error(err))
			sc.Close()
			return
		}
	}
	sc.Close()

	pubNode, err := met.GetOrg(pubID)
	if err != nil {
		met.logger.Debug("this event msg is not for this node", zap.String("source", pubID))
		return
	}

	eventID := fmt.Sprintf("%s,%s,%d", subID, pubID, height)
	datahash, _ := tools.HashKeccak(data)

	emt := &EventMsgTx{
		Listener: subID,
		EventID:  eventID,
		DataHash: datahash,
		Msg:      emsg,
		Time:     time.Now(),
	}
	if _, err := cvtools.TxSign(emt, &privkey); err != nil {
		met.logger.Error("fail to sign EventMsgTx", zap.Error(err))
		return
	}
	txBytes, _ := cvtools.TxToBytes(emt)
	if err := pubNode.GetEngine().BroadcastTx(agtypes.WrapTx(EventMsgTag, txBytes)); err != nil {
		met.logger.Error("broadcast eventMsgTx failed", zap.Error(err))
		return
	}
}

func (met *Metropolis) fetchEvent(addr string, eventID string, hash []byte) ([]byte, error) {
	retryLimit := 3
	retry := 0
	var err error
	var conn net.Conn
	for {
		if retry == retryLimit {
			return nil, errors.Wrap(err, "retrying connect")
		}
		if conn, err = net.Dial("tcp", addr); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
		retry++
	}

	privkey, _ := met.core.GetPrivateKey()
	sconn, err := p2p.MakeSecretConnection(conn, privkey)

	// write subscriberID,publisherID,height
	if _, err := sconn.Write([]byte(eventID)); err != nil {
		sconn.Close()
		return nil, errors.Wrap(err, "fail to write event index")
	}
	var edata []byte
	{
		sconn.SetReadDeadline(time.Now().Add(5 * time.Second))
		rbuf := make([]byte, 2048*10000)
		rn, err := sconn.Read(rbuf)
		if err != nil {
			return nil, errors.Wrap(err, "fail to read event data")
		}
		edata = rbuf[:rn]
	}

	msg := "event has been confirm"
	if _, err := sconn.Write([]byte(msg)); err != nil {
		sconn.Close()
		return nil, errors.Wrap(err, "fail to send msg")
	}

	sconn.Close()

	dhash, _ := tools.HashKeccak(edata)
	if !bytes.Equal(dhash, hash) {
		return nil, fmt.Errorf("wrong data hash, expected: %X, got: %X", hash, dhash)
	}

	return edata, nil
}
