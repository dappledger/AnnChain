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

package ikhofi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	pbtypes "github.com/dappledger/AnnChain/angine/protos/types"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/xlib/def"
	civil "github.com/dappledger/AnnChain/src/chain/node"
)

type LastBlockInfo struct {
	Height  def.INT
	AppHash []byte
}

type IKHOFIApp struct {
	agtypes.BaseApplication

	civil.EventAppBase

	core civil.Core

	logger   *zap.Logger
	datadir  string
	stateMtx sync.Mutex

	Config      *viper.Viper
	AngineHooks agtypes.Hooks
}

type Response struct {
	Code    int    "code"
	Value   string "value"
	Message string "message"
}

type StartParams struct {
	ConfigPath   string
	StateHashHex string
}

var (
	ikhofiSigner = DawnSigner{}

	SystemContractId                  = "system"
	SystemDeployMethod                = "deploy"
	SystemUpgradeMethod               = "upgrade"
	SystemQueryContractIdExits        = "contract"
	SystemQueryEventFilterById        = "eventFilterById"
	SystemQueryEventFilterByIdAndType = "eventFilterByIdAndType"

	serverUrl = ""
)

func NewIKHOFIApp(logger *zap.Logger, config *viper.Viper) (*IKHOFIApp, error) {

	serverUrl = config.GetString("ikhofi_addr")
	if serverUrl == "" {
		return nil, errors.Wrap(errors.Errorf("app miss configuration ikhofi_addr"), "app error")
	}

	app := IKHOFIApp{
		logger:  logger,
		datadir: config.GetString("db_dir"),

		Config:       config,
		EventAppBase: civil.NewEventAppBase(zap.NewExample(), config.GetString("cosi_laddr")),
	}

	if err := app.BaseApplication.InitBaseApplication("ikhofi", app.datadir); err != nil {
		return nil, errors.Wrap(err, "app error")
	}

	app.AngineHooks = agtypes.Hooks{
		OnNewRound: agtypes.NewHook(app.OnNewRound),
		OnCommit:   agtypes.NewHook(app.OnCommit),
		// OnPrevote:  agtypes.NewHook(app.OnPrevote),
		OnExecute: agtypes.NewHook(app.OnExecute),
	}

	return &app, nil
}

func (app *IKHOFIApp) SetCore(c civil.Core) {
	app.core = c
	app.EventAppBase.SetCore(c)
}

func (app *IKHOFIApp) HandleEvent(eventData civil.EventData, notification *civil.EventNotificationTx) {
	bs, err := hex.DecodeString(eventData["from"].(string))
	if err != nil {
		//fmt.Println(err)
		panic(err)
	}

	addr := common.BytesToAddress(bs)
	// make NewTransaction
	score := eventData["score"].(string)
	tx := NewTransaction(addr, "Chicken", "buyChickenByScores", []string{score}, nil)

	txHash := GetHash(tx.Transaction2PbTmp())

	// encode tx to bytes
	txpb := &TransactionPb{
		From:     tx.From[:],
		To:       tx.To,
		Method:   tx.Method,
		Args:     tx.Args,
		Bytecode: tx.ByteCode,
		Nonce:    tx.Nonce,
		Hash:     txHash[:],
	}

	b, err := proto.Marshal(txpb)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%X\n", b)
	app.core.GetEngine().BroadcastTx(b)
}

func (app *IKHOFIApp) ExecuteTx(height def.INT, block *agtypes.BlockCache, bs []byte) (validtx []byte, err error) {
	txpb := &TransactionPb{}
	if err = proto.Unmarshal(bs, txpb); err != nil {
		return
	}
	tx := Pb2Transaction(txpb)
	txInfo := &TxInfo{
		LastCommitHash: block.Header.LastCommitHash,
		BlockHeight:    height,
		TxFrom:         common.Bytes2Hex(tx.From[:]),
		BlockTime:      block.Header.Time,
		TxHash:         tx.Hash[:],
	}
	executionData := &Execution{
		Version:  1,
		Id:       tx.To,
		Method:   tx.Method,
		TxInfo:   txInfo,
		ByteCode: tx.ByteCode,
	}
	if len(tx.Args) > 0 {
		executionData.Args = tx.Args
	}
	executionDataBytes, _ := proto.Marshal(executionData)
	result, err := app.post("execute", "text/plain", executionDataBytes)
	if err != nil {
		return
	}

	executeRes := &Result{}
	err = proto.Unmarshal(result, executeRes)
	if err != nil {
		return
	}

	if executeRes.Code != 0 {
		err = fmt.Errorf("execute error:" + strconv.Itoa(int(executeRes.Code)))
		app.logger.Error("ikhofi execute tx error", zap.Error(err))
		return
	}

	return result, nil
}

func (app *IKHOFIApp) Start() (err error) {
	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}

	trieRoot := ""
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if err != nil {
		app.logger.Error("fail to load last block", zap.Error(err))
		return
	}
	trieRoot = common.Bytes2Hex(lastBlock.AppHash)

	path := app.Config.GetString("ikhofi_config")
	if path[0:1] != "/" {
		pwd, _ := os.Getwd()
		path = filepath.Join(pwd, path)
	}

	startParams := StartParams{
		path,
		trieRoot,
	}

	startParamsB, err := json.Marshal(startParams)
	if err != nil {
		app.logger.Error("json marshal error", zap.Error(err))
		return
	}
	resultB, err := app.post("start", "application/json", startParamsB)
	result, _ := strconv.ParseBool(string(resultB))
	if !result {
		app.Stop()
		return errors.Wrap(errors.Errorf("ikhofi jvm start error"), "[IKHOFIApp Start]")
	}

	if _, err := app.EventAppBase.Start(); err != nil {
		app.Stop()
		return errors.Wrap(err, "[IKHOFIApp Start]")
	}

	return nil
}

func (app *IKHOFIApp) Stop() {
	app.get("stop")
}

func (app *IKHOFIApp) GetAngineHooks() agtypes.Hooks {
	return app.AngineHooks
}

func (app *IKHOFIApp) CompatibleWithAngine() {}

func (app *IKHOFIApp) SupportCoSiTx() {}

func (app *IKHOFIApp) CheckTx(bs []byte) error {
	txpb := &TransactionPb{}
	if err := proto.Unmarshal(bs, txpb); err != nil {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, err.Error())
	}
	tx := Pb2Transaction(txpb)

	from, _ := Sender(ikhofiSigner, tx)

	app.stateMtx.Lock()
	defer app.stateMtx.Unlock()

	// check whether calculate address is equal to tx.to
	if from != tx.From {
		return fmt.Errorf("address from signed tx is not same to tx from address")
	}

	return nil
}

func (app *IKHOFIApp) Query(query []byte) agtypes.Result {
	txpb := &TransactionPb{}
	if err := proto.Unmarshal(query, txpb); err != nil {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidInput, err.Error())
	}
	tx := Pb2Transaction(txpb)

	queryData := &Query{
		Version: 1,
		Id:      tx.To,
		Method:  tx.Method,
	}
	if len(tx.Args) > 0 {
		queryData.Args = tx.Args
	}

	queryDataBytes, _ := proto.Marshal(queryData)
	res, err := app.post("query", "text/plain", queryDataBytes)
	if err != nil {
		return agtypes.NewError(pbtypes.CodeType_BaseInvalidOutput, err.Error())
	}

	return agtypes.NewResultOK(res, "")
}

func (app *IKHOFIApp) Info() (resInfo agtypes.ResultInfo) {
	lb := &LastBlockInfo{
		AppHash: make([]byte, 0),
		Height:  0,
	}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}
	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "alpha 0.6"
	resInfo.Data = "default app with ikhofi-0.6, cosi and eventtx"
	return
}

func (app *IKHOFIApp) OnNewRound(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	//TODO
	return agtypes.NewRoundResult{}, nil
}

func (app *IKHOFIApp) OnPrevote(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	//TODO
	return nil, nil
}

func (app *IKHOFIApp) OnExecute(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	var (
		res agtypes.ExecuteResult
		err error
	)

	app.stateMtx.Lock()

	for _, tx := range block.Data.Txs {
		if _, err := app.ExecuteTx(height, block, tx); err != nil {
			res.InvalidTxs = append(res.InvalidTxs, agtypes.ExecuteInvalidTx{Bytes: tx, Error: err})
		} else {
			res.ValidTxs = append(res.ValidTxs, tx)
		}
	}

	app.stateMtx.Unlock()

	return res, err
}

func (app *IKHOFIApp) OnCommit(height, round def.INT, block *agtypes.BlockCache) (interface{}, error) {
	appHash, _ := app.get("commit")

	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash})

	return agtypes.CommitResult{
		AppHash: appHash,
	}, nil
}

func (app *IKHOFIApp) get(method string) (result []byte, err error) {
	url := serverUrl + "/" + method
	resp, err := http.Get(url)
	if err != nil {
		app.logger.Error("http get error", zap.Error(err))
		return
	}

	defer resp.Body.Close()
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		app.logger.Error("io read error", zap.Error(err))
		return
	}

	return
}

func (app *IKHOFIApp) post(method, contentType string, data []byte) (result []byte, err error) {
	url := serverUrl + "/" + method
	resp, err := http.Post(url, contentType, bytes.NewBuffer(data))
	if err != nil {
		app.logger.Error("http post error", zap.Error(err))
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		app.logger.Error("request status: "+method+" "+string(resp.StatusCode), zap.Error(err))
		return
	}

	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		app.logger.Error("io read error", zap.Error(err))
		return
	}

	return
}
