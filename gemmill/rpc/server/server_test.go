package rpcserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/rpc/client"
	gtypes "github.com/dappledger/AnnChain/gemmill/rpc/types"
	"github.com/dappledger/AnnChain/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setAuditLogger(logWriter io.Writer) {
	zapEncodeConfig := zap.NewProductionEncoderConfig()
	jsonEncoder := zapcore.NewJSONEncoder(zapEncodeConfig)

	w := zapcore.AddSync(logWriter)
	core := zapcore.NewCore(
		jsonEncoder,
		w,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel
		}),
	)
	logger := zap.New(core)
	log.SetAuditLog(logger)
}

func mustNewRequest(t *testing.T, path string , body interface{}) *http.Request{
	uri := "http://127.0.0.1:54329"
	var data []byte
	if body!= nil {
		data,_ = gtypes.JsonBytes(body)
	}
	req,err := http.NewRequest("POST",uri+"/"+path,bytes.NewReader(data))
	req.RemoteAddr ="10.2.3.4"
	assert.NoError(t,err)
	return req
}

func TestAuditLog(t *testing.T) {
	logData := bytes.NewBuffer(nil)
	setAuditLogger(logData)
	var okResponse string
    handleAndCheck(t,handleOk,mustNewRequest(t,"ok",nil),&okResponse,nil)
	assert.Equal(t, okResponse, "ok")
	checkLog(t,logData)
	var echoResponse responseData
	var expectedResp = responseData{
		Data:   []byte( "\"this is a test\""),
		Message: "hello",
		Code:    200,
	}
	reqData:= "this is a test"
	handleAndCheck(t,handleEcho,mustNewRequest(t,"echo",&reqData),&echoResponse,nil)
	assert.Equal(t, echoResponse, expectedResp)
	checkLog(t,logData)
	handleAndCheck(t,handleErr,mustNewRequest(t,"err",nil),nil,errors.New("this is an error"))
	checkLog(t,logData)
}

func checkLog (t *testing.T, r *bytes.Buffer) {
	firstLine,err := r.ReadBytes('\n')
	assert.NoError(t,err)
	secondLine,err := r.ReadBytes('\n')
	assert.NoError(t,err)
	var request = make( map[string]interface{})
	var response = make(map[string]interface{})
	err  = json.Unmarshal(firstLine, &request)
	assert.NoError(t,err)
	err  = json.Unmarshal(secondLine, &response)
	assert.NoError(t,err)
	assert.Equal(t,request["client_ip"],"10.2.3.4")
	assert.Equal(t,response["status"],float64(200))
	assert.Equal(t,request["trace_id"],response["trace_id"])
	traceId ,err := utils.TraceIDFromString(fmt.Sprintf("%v",request["trace_id"]))
	assert.NoError(t,err)
	timeStamp:= traceId.Timestamp()
	reqDuration,err := time.ParseDuration(fmt.Sprintf("%v",response["req_duration"]))
	assert.NoError(t,err)
	assert.True(t, 0<=reqDuration && reqDuration < time.Second*10)
	latency:= time.Now().Sub(timeStamp)
	assert.True(t, 0<=latency && latency< time.Second*20)
}

func handleAndCheck(t *testing.T, handler http.HandlerFunc, req *http.Request, actual interface{}, expectedErr error) {
	rw :=httptest.NewRecorder()
	h:= RecoverAndLogHandler(handler)
	h.ServeHTTP(rw,req)
	assert.Equal(t, rw.Code, 200)
	var res gtypes.RPCResponse
	err := json.NewDecoder(rw.Body).Decode(&res)
	assert.NoError(t, err)
	if expectedErr == nil {
		assert.Equal(t, "", res.Error)
	} else {
		assert.Equal(t, expectedErr, errors.New(res.Error))
	}
	if actual == nil {
		assert.Nil(t, res.Result)
	} else {
		assert.NotNil(t, res.Result)
		_, err = rpcclient.ReadJSONObjectPtr(actual, *res.Result)
		assert.NoError(t, err)
	}
	return
}



func handleOk( w http.ResponseWriter ,req*http.Request) {
	var str = "ok"
	WriteRPCResponseHTTP(w,gtypes.NewRPCResponse("",&str,""))
}

type responseData struct {
	Data    []byte `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleEcho( w http.ResponseWriter ,req*http.Request) {
	var body []byte
	if req.Body!=nil {
		body,_ = ioutil.ReadAll(req.Body)
	}
	res:=  &responseData {
		Data:    body,
		Message: "hello",
		Code:    200,
	}
	WriteRPCResponseHTTP(w,gtypes.NewRPCResponse("",res,""))
}

func handleErr(w http.ResponseWriter ,req*http.Request) {
	WriteRPCResponseHTTP(w,gtypes.NewRPCResponse("",nil,"this is an error"))
}
