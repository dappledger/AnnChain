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
	"time"

	"github.com/dappledger/AnnChain/module/xlib/def"
	civiltypes "github.com/dappledger/AnnChain/src/types"
)

var (
	EventTag             = []byte{'e', 'v', 't'}
	EventRequestTag      = append(EventTag, 0x01)
	EventSubscribeTag    = append(EventTag, 0x02)
	EventNotificationTag = append(EventTag, 0x03)
	EventUnsubscribeTag  = append(EventTag, 0x04)
	EventConfirmTag      = append(EventTag, 0x05)
	EventMsgTag          = append(EventTag, 0x06)
	EventUploadCodeTag   = append(EventTag, 0x07)
)

type EventRequestTx struct {
	civiltypes.CivilTx

	Source       string    `json:"source"`
	Listener     string    `json:"listener"`
	Time         time.Time `json:"time"`
	SourceHash   []byte    `json:"source_hash"`
	ListenerHash []byte    `json:"listener_hash"`
}

type EventSubscribeTx struct {
	civiltypes.CivilTx

	Source      string    `json:"source"`
	Threshold   int       `json:"threshold"`
	TxHash      []byte    `json:"txhash"`
	Time        time.Time `json:"time"`
	SignData    []byte    `json:"signdata"`
	CoSignature []byte    `json:"cosignature"`
}

type EventNotificationTx struct {
	civiltypes.CivilTx

	Listener  string    `json:"listener"`
	Source    string    `json:"source"`
	Height    def.INT   `json:"height"`
	DataHash  []byte    `json:"datahash"`
	RelatedTx []byte    `json:"relatedtx"`
	Time      time.Time `json:"time"`
}

type EventUnsubscribeTx struct {
	civiltypes.CivilTx

	Source   string    `json:"source"`
	Listener string    `json:"listener"`
	Proof    []byte    `json:"proof"`
	Time     time.Time `json:"time"`
}

type EventUploadCodeTx struct {
	civiltypes.CivilTx
	Source string    `json:"source"`
	Owner  string    `json:"owner"`
	Code   string    `json:"code"`
	Time   time.Time `json:"time"`
}

type EventConfirmTx struct {
	civiltypes.CivilTx
	Source   string    `json:"source"`
	EventID  string    `json:"eventid"`
	DataHash []byte    `json:"datahash"`
	TxHash   []byte    `json:"txhash"`
	Time     time.Time `json:"time"`
}

type EventMsgTx struct {
	civiltypes.CivilTx
	Listener string    `json:"listener"`
	EventID  string    `json:"eventid"`
	DataHash []byte    `json:"datahash"`
	Msg      []byte    `json:"msg"`
	Time     time.Time `json:"time"`
}

func IsEventTx(tx []byte) bool {
	return bytes.HasPrefix(tx, EventTag)
}

func IsEventRequestTx(tx []byte) bool {
	return bytes.Equal(EventRequestTag, tx[:4])
}

func IsEventSubscribeTx(tx []byte) bool {
	return bytes.Equal(EventSubscribeTag, tx[:4])
}

func IsEventNotificationTx(tx []byte) bool {
	return bytes.Equal(EventNotificationTag, tx[:4])
}

func IsEventUnsubscribeTx(tx []byte) bool {
	return bytes.Equal(EventUnsubscribeTag, tx[:4])
}

func IsEventUploadCodeTx(tx []byte) bool {
	return bytes.Equal(EventUploadCodeTag, tx[:4])
}
