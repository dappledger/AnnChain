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
	"fmt"
	"sync/atomic"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/dappledger/AnnChain/angine"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	//	client "github.com/dappledger/AnnChain/module/lib/go-rpc/client"
)

const (
	RPCCollectSpecialVotes uint8 = iota
)

// OrgNode implements types.Core
type OrgNode struct {
	running int64

	logger *zap.Logger

	appname string

	Superior    Superior
	Angine      *angine.Angine
	AngineTune  *angine.Tunes
	Application agtypes.Application
	GenesisDoc  *agtypes.GenesisDoc
}

func NewOrgNode(logger *zap.Logger, appName string, conf *viper.Viper, metro *Metropolis) *OrgNode {
	defer func() {
		// in case App constructor calls panic
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if _, ok := Apps[appName]; !ok {
		return nil
	}

	tune := &angine.Tunes{Conf: conf}
	ang := angine.NewAngine(logger, tune)
	if ang == nil {
		logger.Error("fail to new Angine")
		return nil
	}
	app, err := Apps[appName](logger, conf, ang.PrivValidator().GetPrivKey())
	if err != nil {
		ang.Stop()
		logger.Error("fail to create app", zap.String("appname:", appName), zap.Error(err))
		return nil
	}
	if err := ang.ConnectApp(app); err != nil {
		ang.Stop()
		logger.Error("fail to connect app", zap.Error(err))
		return nil
	}

	orgnode := &OrgNode{
		running: 0,
		logger:  logger,
		appname: appName,

		Superior:    metro,
		Application: app,
		Angine:      ang,
		AngineTune:  tune,
		GenesisDoc:  ang.Genesis(),
	}
	app.SetCore(orgnode)

	return orgnode
}

func (o *OrgNode) GetAppName() string {
	return o.appname
}

func (o *OrgNode) Start() error {
	if atomic.CompareAndSwapInt64(&o.running, 0, 1) {
		if err := o.Application.Start(); err != nil {
			o.Angine.Stop()
			return err
		}
		if err := o.Angine.Start(); err != nil {
			o.Application.Stop()
			o.Angine.Stop()
			return err
		}
		for o.Angine.Genesis() == nil {
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	}
	return fmt.Errorf("already started")
}

func (o *OrgNode) Stop() bool {
	if atomic.CompareAndSwapInt64(&o.running, 1, 0) {
		o.Angine.Stop()
		o.Application.Stop()
		return true
	}
	return false
}

func (o *OrgNode) IsRunning() bool {
	return atomic.LoadInt64(&o.running) == 1
}

func (o *OrgNode) GetEngine() Engine {
	return o.Angine
}

func (o *OrgNode) GetPublicKey() (r crypto.PubKeyEd25519, b bool) {
	var pr *crypto.PubKeyEd25519
	pr, b = o.Angine.PrivValidator().GetPubKey().(*crypto.PubKeyEd25519)
	r = *pr
	return
}

func (o *OrgNode) GetPrivateKey() (r crypto.PrivKeyEd25519, b bool) {
	var pr *crypto.PrivKeyEd25519
	pr, b = o.Angine.PrivValidator().GetPrivKey().(*crypto.PrivKeyEd25519)
	r = *pr
	return
}

func (o *OrgNode) IsValidator() bool {
	_, vals := o.Angine.GetValidators()
	for i := range vals.Validators {
		if vals.Validators[i].GetPubKey().KeyString() == o.Angine.PrivValidator().GetPubKey().KeyString() {
			return true
		}
	}
	return false
}

func (o *OrgNode) GetChainID() string {
	return o.Angine.Genesis().ChainID
}

func (o *OrgNode) BroadcastTxSuperior(tx []byte) error {
	return o.Superior.BroadcastTx(tx)
}

func (o *OrgNode) PublishEvent(from string, block *agtypes.BlockCache, data []EventData, txhash []byte) error {
	return o.Superior.PublishEvent(from, block, data, txhash)
}

func (o *OrgNode) CodeExists(codehash []byte) bool {
	return o.Superior.CodeExists(codehash)
}
