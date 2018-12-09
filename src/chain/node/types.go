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
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/module/xlib/def"
	"github.com/dappledger/AnnChain/src/tools"
)

type (
	EventData map[string]interface{}

	eventItem struct {
		Name  string
		Value interface{}
	}

	event []eventItem
)

func newEventItem(n string, v interface{}) eventItem {
	return eventItem{
		Name:  n,
		Value: v,
	}
}

func EncodeEventData(data EventData) ([]byte, error) {
	// buf := new(bytes.Buffer)
	// encoder := gob.NewEncoder(buf)
	// if err := encoder.Encode(data); err != nil {
	// 	return nil, errors.Wrap(err, "[encodeEventData]")
	// }

	// return buf.Bytes(), nil

	sortedK := make([]string, 0)
	for k := range data {
		sortedK = append(sortedK, k)
	}
	sort.Strings(sortedK)

	ev := make(event, len(sortedK))
	for i, k := range sortedK {
		ev[i] = newEventItem(k, data[k])
	}

	return json.Marshal(ev)
}

func DecodeEventData(bs []byte) (EventData, error) {
	// data := make(map[string]interface{})
	// if err := gob.NewDecoder(bytes.NewReader(bs)).Decode(&data); err != nil {
	// 	return nil, errors.Wrap(err, "[decodeEventData]")
	// }
	// return data, nil
	ev := make(event, 0)
	err := json.Unmarshal(bs, &ev)
	if err != nil {
		return nil, err
	}

	data := make(EventData)
	for _, v := range ev {
		data[v.Name] = v.Value
	}

	return data, err
}

type (
	// Application embeds types.Application, defines application interface in Civilization
	Application interface {
		agtypes.Application
		SetCore(Core)
		GetAttributes() AppAttributes
	}

	// Core defines the interface at which an application sees its containing organization
	Core interface {
		Publisher

		IsValidator() bool
		GetPublicKey() (crypto.PubKeyEd25519, bool)
		GetPrivateKey() (crypto.PrivKeyEd25519, bool)
		GetChainID() string
		GetEngine() Engine
		BroadcastTxSuperior([]byte) error
	}

	// Engine defines the consensus engine
	Engine interface {
		GetBlock(def.INT) (*agtypes.BlockCache, *pbtypes.BlockMeta, error)
		GetBlockMeta(def.INT) (*pbtypes.BlockMeta, error)
		GetValidators() (def.INT, *agtypes.ValidatorSet)
		PrivValidator() *agtypes.PrivValidator
		BroadcastTx([]byte) error
		Query(byte, []byte) (interface{}, error)
	}

	// Superior defines the application on the upper level, e.g. Metropolis
	Superior interface {
		Publisher
		Broadcaster
	}

	// Broadcaster means we can deliver tx in application
	Broadcaster interface {
		BroadcastTx([]byte) error
	}

	// Publisher means that we can publish events
	Publisher interface {
		// PublishEvent
		// if data is neither tx nor []tx, the related tx hash should be given accordingly
		PublishEvent(from string, block *agtypes.BlockCache, data []EventData, txhash []byte) error
		CodeExists([]byte) bool
	}

	// Serializable transforms to bytes
	Serializable interface {
		ToBytes() ([]byte, error)
	}

	// Unserializable transforms from bytes
	Unserializable interface {
		FromBytes(bs []byte)
	}

	// Hashable aliases Serializable
	Hashable interface {
		Serializable
	}

	// EventSubscriber defines interfaces for being compatible with event system
	EventSubscriber interface {
		// IncomingEvent returns App's decision about whether to fetch this new event
		IncomingEvent(from string, height def.INT) bool

		ConfirmEvent(tx *EventNotificationTx) error

		// HandleEvent will process the event, you can do whatever you like. This will be running in its own routine
		HandleEvent(data EventData, tx *EventNotificationTx)
	}

	// EventRequestHandle defines just one empty method, SupportCoSiTx, as a token
	// Event system will need the Ed25519 CoSi
	EventRequestHandle interface {
		SupportCoSiTx()
	}

	// EventApp defines what an app need to implement to support cross-org event communication
	EventApp interface {
		EventRequestHandle
		EventSubscriber
	}

	// AppMaker is the signature for functions which take charge of create new instance of applications
	AppMaker func(*zap.Logger, *viper.Viper, crypto.PrivKey) (Application, error)
)

// AppAttributes is just a type alias
type AppAttributes = map[string]string

type IMetropolisApp interface {
	GetAttribute(string) (string, bool)
	GetAttributes() AppAttributes
	SetAttributes(AppAttributes)
	PushAttribute(string, string)
	AttributeExists(string) bool
}

// MetropolisAppBase is the base struct every app in annchain should embed
type MetropolisAppBase struct {
	attributes AppAttributes
	core       Core
}

func NewMetropolisAppBase() MetropolisAppBase {
	return MetropolisAppBase{
		attributes: make(AppAttributes),
	}
}

func (app *MetropolisAppBase) SetCore(c Core) {
	app.core = c
}

func (app *MetropolisAppBase) Start() error {
	return nil
}

func (app *MetropolisAppBase) Stop() error {
	return nil
}

func (app *MetropolisAppBase) GetAttribute(key string) (val string, exists bool) {
	if len(app.attributes) == 0 {
		return
	}

	val, exists = app.attributes[key]
	return
}

func (app *MetropolisAppBase) GetAttributes() AppAttributes {
	return app.attributes
}

func (app *MetropolisAppBase) SetAttributes(attrs AppAttributes) {
	app.attributes = attrs
}

func (app *MetropolisAppBase) PushAttribute(key, val string) {
	if app.attributes == nil {
		app.attributes = make(AppAttributes)
	}

	app.attributes[key] = val
}

func (app *MetropolisAppBase) AttributeExists(key string) bool {
	if len(app.attributes) == 0 {
		return false
	}

	_, ok := app.attributes[key]
	return ok
}

// EventAppBase gives an app the ability to deal with the event system provided by metropolis
type EventAppBase struct {
	MetropolisAppBase
	CoSiAppBase

	core Core

	logger   *zap.Logger
	cosiAddr string
}

func GetCoSiAddress(attr AppAttributes) string {
	return attr["cosi_external_address"]
}

func NewEventAppBase(l *zap.Logger, cosiAddr string) EventAppBase {
	return EventAppBase{
		logger:   l,
		cosiAddr: cosiAddr,

		MetropolisAppBase: NewMetropolisAppBase(),
		CoSiAppBase:       NewCoSiAppBase(cosiAddr),
	}
}

func (app *EventAppBase) SetCore(core Core) {
	app.core = core
	app.CoSiAppBase.SetCore(core)
	app.MetropolisAppBase.SetCore(core)
}

func (app *EventAppBase) SetCoSiAddr(addr string) {
	app.cosiAddr = addr
	app.CoSiAppBase.cosiAddr = addr
}

func (app *EventAppBase) Start() (string, error) {
	if app.cosiAddr == "" {
		return "", errors.Wrap(errors.Errorf("CoSi address is missing"), "[EventAppBase Start]")
	}

	if err := app.MetropolisAppBase.Start(); err != nil {
		return "", errors.Wrap(err, "[EventAppBase start metropolis app base error]")
	}
	external, err := app.Setup()
	if err != nil {
		return "", errors.Wrap(err, "[EventAppBase Start]")
	}

	app.MetropolisAppBase.PushAttribute("cosi_external_address", external)

	return external, nil
}

func (app *EventAppBase) Stop() error {
	app.MetropolisAppBase.Stop()
	return nil
}

// SupportCoSiTx is just an empty token
func (app *EventAppBase) SupportCoSiTx() {}

func (app *EventAppBase) PublishEvent(data []EventData, block *agtypes.BlockCache) error {
	return app.core.PublishEvent(app.core.GetChainID(), block, data, nil)
}

// IncomingEvent is set true by default
func (app *EventAppBase) IncomingEvent(_ string, _ def.INT) bool {
	return true
}

/*
 * TO BE IMPLEMENTED
 */
// func (app *EventAppBase) HandleEvent(eventData civil.EventData, notification *EventNotificationTx) {}

// ConfirmEvent broadcasts an confirm tx on the organization
func (app *EventAppBase) ConfirmEvent(tx *EventNotificationTx) error {
	// pubkey, _ := app.Core.GetPublicKey()
	privkey, _ := app.core.GetPrivateKey()
	eventID := fmt.Sprintf("%s,%s,%d", tx.Listener, tx.Source, tx.Height)

	ftx := &EventConfirmTx{
		Source:   tx.Source,
		EventID:  eventID,
		DataHash: tx.DataHash,
		Time:     time.Now(),
	}
	ftx.TxHash, _ = tools.TxHash(tx)
	if _, err := tools.TxSign(ftx, &privkey); err != nil {
		return errors.Wrap(err, "[EventAppBase ConfirmEvent]")
	}
	txBytes, _ := tools.TxToBytes(ftx)
	if err := app.core.GetEngine().BroadcastTx(agtypes.WrapTx(EventConfirmTag, txBytes)); err != nil {
		return errors.Wrap(err, "[EventAppBase ConfirmEvent]")
	}

	return nil
}

func (app *EventAppBase) CheckTx(bs []byte) error {
	if !IsCoSiTx(bs) {
		return nil
	}

	if err := app.CoSiAppBase.CheckTx(bs); err != nil {
		return errors.Wrap(err, "[EventAppBase CheckTx]")
	}

	txBytes := agtypes.UnwrapTx(bs)
	cositx := &CoSiTx{}
	tools.TxFromBytes(txBytes, cositx)
	if cositx.Type == "eventrequest" {
		data := bytes.Split(cositx.Data, []byte("|<>|"))
		if len(data) != 2 {
			return errors.Wrap(errors.Errorf("malformed event data"), "[EventAppBase CheckTx]")
		}
		if !app.core.CodeExists(data[0]) {
			return errors.Wrap(errors.Errorf("event code doesn't exist"), "[EventAppBase CheckTx]")
		}
	}

	return nil
}

func (app *EventAppBase) ExecuteTx(bs []byte, validators *agtypes.ValidatorSet) error {
	if !IsCoSiTx(bs) {
		return nil
	}

	txBytes := agtypes.UnwrapTx(bs)
	cositx := &CoSiTx{}
	if err := tools.TxFromBytes(txBytes, cositx); err != nil {
		return errors.Wrap(err, "[EventAppBase ExecuteTx]")
	}

	if cositx.Type != "eventrequest" {
		return nil
	}

	// non-validator can't participate cosi procedure
	if !app.core.IsValidator() {
		return nil
	}

	threshold := validators.Size()*2/3 + 1
	pubkey, _ := app.core.GetPublicKey()
	privkey, _ := app.core.GetPrivateKey()
	cm := NewCoSiModule(app.logger, privkey, validators)
	if !bytes.Equal(cositx.Leader, pubkey[:]) {
		// potential follower

		// if we disapprove this cosi, just pass an empty []byte instead of cositx.Data,
		// so we can do all kinds of checks on the cositx.Data as we want

		data := bytes.Split(cositx.Data, []byte("|<>|"))
		if app.core.CodeExists(data[0]) {
			if err := cm.FollowCoSign(cositx.LeaderAddr, cositx.Data); err != nil {
				return errors.Wrap(err, "[EventAppBase ExecuteTx]")
			}
		} else {
			if err := cm.FollowCoSign(cositx.LeaderAddr, []byte{}); err != nil {
				return errors.Wrap(err, "[EventAppBase ExecuteTx]")
			}
		}

	} else {
		cosignature, err := cm.LeadCoSign(app.cosiAddr, cositx.Data)
		if err != nil {
			fmt.Println(errors.Wrap(err, "[EventAppBase ExecuteTx]"))
			return errors.Wrap(err, "[EventAppBase ExecuteTx]")
		}

		// civil.EventSubscribeTx receivers will verify the cosi signature one more time.
		estx := &EventSubscribeTx{
			Source:      app.core.GetChainID(),
			Threshold:   threshold,
			TxHash:      cositx.TxHash,
			SignData:    cositx.Data,
			CoSignature: cosignature,
		}
		tools.TxSign(estx, &privkey)
		txBytes, _ := tools.TxToBytes(estx)
		fmt.Println("broadcast subscription")
		if err := app.core.BroadcastTxSuperior(agtypes.WrapTx(EventSubscribeTag, txBytes)); err != nil {
			fmt.Println(errors.Wrap(err, "[EventAppBase ExecuteTx]"))
			return errors.Wrap(err, "[EventAppBase ExecuteTx]")
		}
	}

	return nil
}

func (app *EventAppBase) Setup() (string, error) {
	external, err := app.CoSiAppBase.Setup()
	if err != nil {
		return "", errors.Wrap(err, "[EventAppBase Setup]")
	}
	app.cosiAddr = app.CoSiAppBase.cosiAddr

	return external, err
}

// CoSiAppBase is the base app for CoSi capability
type CoSiAppBase struct {
	core     Core
	cosiAddr string
}

func NewCoSiAppBase(addr string) CoSiAppBase {
	return CoSiAppBase{
		cosiAddr: addr,
	}
}

func (app *CoSiAppBase) SetCore(core Core) {
	app.core = core
}

func (app *CoSiAppBase) CheckTx(bs []byte) error {
	if !IsCoSiTx(bs) {
		return nil
	}

	txBytes := agtypes.UnwrapTx(bs)
	cositx := &CoSiTx{}
	if err := tools.TxFromBytes(txBytes, cositx); err != nil {
		return errors.Wrap(err, "[CoSiAppBase CheckTx]")
	}
	if v, err := tools.TxVerifySignature(cositx); err != nil {
		return errors.Wrap(err, "[CoSiAppBase CheckTx]")
	} else if !v {
		return errors.Wrap(errors.Errorf("signature verification failed"), "[CoSiAppBase CheckTx]")
	}

	return nil
}

func (app *CoSiAppBase) ExecuteTx(bs []byte) error {
	return nil
}

func (app *CoSiAppBase) Setup() (string, error) {
	protocol, address := tools.ProtocolAndAddress(app.cosiAddr)
	lAddrIP, lAddrPort, err := net.SplitHostPort(address)
	if err != nil {
		return "", errors.Wrap(err, "[CoSiAppBase Setup]")
	}
	listener, err := net.Listen(protocol, address)
	if err != nil {
		return "", errors.Wrap(err, "[CoSiAppBase Setup]")
	}
	cosiExternal, err := tools.DetermineExternalAddress(listener, lAddrIP, lAddrPort, true)
	if err != nil {
		listener.Close()
		return "", errors.Wrap(errors.Wrap(err, "fail to determine the event external address"), "[CoSiAppBase Setup]")

	}
	listener.Close()

	return cosiExternal, nil
}
