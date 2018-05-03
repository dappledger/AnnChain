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

package plugin

import (
	"time"

	"github.com/pkg/errors"

	"github.com/dappledger/AnnChain/angine/types"
	dbm "github.com/dappledger/AnnChain/module/lib/go-db"
)

type QueryCachePlugin struct {
	db          dbm.DB
	eventSwitch types.EventSwitch
}

func (qc *QueryCachePlugin) Init(p *InitParams) {
	qc.db = p.DB
}

func (qc *QueryCachePlugin) Reload(p *ReloadParams) {

}

func (qc *QueryCachePlugin) CheckTx(tx []byte) (bool, error) {
	return false, nil
}

func (qc *QueryCachePlugin) DeliverTx(tx []byte, i int) (bool, error) {
	return false, nil
}

func (qc *QueryCachePlugin) BeginBlock(p *BeginBlockParams) (*BeginBlockReturns, error) {
	return nil, nil
}

func (qc *QueryCachePlugin) EndBlock(p *EndBlockParams) (*EndBlockReturns, error) {
	return nil, nil
}

func (qc *QueryCachePlugin) ExecBlock(p *ExecBlockParams) (*ExecBlockReturns, error) {
	batch := qc.db.NewBatch()
	bheader := p.Block.Header
	queryBlockInfo, err := (&types.TxExecutionResult{
		Height:        bheader.Height,
		BlockHash:     p.Block.Hash(),
		BlockTime:     time.Unix(0, bheader.Time),
		ValidatorHash: bheader.ValidatorsHash,
	}).ToBytes()
	if err != nil {
		return nil, errors.Errorf("[QueryCachePlugin ExecBlock] fail to serialize execution result: %v", err)
	}

	for _, tx := range p.ValidTxs {
		batch.Set(tx.Hash(), queryBlockInfo)
	}

	batch.Write()

	return nil, nil
}

func (qc *QueryCachePlugin) Reset() {}

func (qc *QueryCachePlugin) Stop() {
	qc.db.Close()
}

func (qc *QueryCachePlugin) ExecutionResult(txHash []byte) (*types.TxExecutionResult, error) {
	item := qc.db.Get(txHash)
	if len(item) == 0 {
		return nil, errors.Errorf("no execution result for %v", txHash)
	}
	info := &types.TxExecutionResult{}
	if err := info.FromBytes(item); err != nil {
		return nil, err
	}
	return info, nil
}

func (qc *QueryCachePlugin) SetEventSwitch(sw types.EventSwitch) {
	qc.eventSwitch = sw
}
