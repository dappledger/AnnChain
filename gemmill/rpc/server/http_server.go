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

// Commons for HTTP handling
package rpcserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dappledger/AnnChain/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	gcmn "github.com/dappledger/AnnChain/gemmill/modules/go-common"
	log "github.com/dappledger/AnnChain/gemmill/modules/go-log"
	gtypes "github.com/dappledger/AnnChain/gemmill/rpc/types"
)

const MaxAuditLogContentSize = 50

func StartHTTPServer(listenAddr string, handler http.Handler) (listener net.Listener, err error) {
	// listenAddr should be fully formed including tcp:// or unix:// prefix
	var proto, addr string
	parts := strings.SplitN(listenAddr, "://", 2)
	if len(parts) != 2 {
		log.Warn("WARNING (go-rpc): Please use fully formed listening addresses, including the tcp:// or unix:// prefix")
		// we used to allow addrs without tcp/unix prefix by checking for a colon
		// TODO: Deprecate
		proto = gtypes.SocketType(listenAddr)
		addr = listenAddr
		// return nil, fmt.Errorf("Invalid listener address %s", lisenAddr)
	} else {
		proto, addr = parts[0], parts[1]
	}

	log.Infof("Starting RPC HTTP server on %s socket %v", proto, addr)
	listener, err = net.Listen(proto, addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to listen to %v: %v", listenAddr, err)
	}

	go func() {
		// res := http.Serve(
		// 	listener,
		// 	RecoverAndLogHandler(handler),
		// )
		ser := &http.Server{Handler: RecoverAndLogHandler(handler), ReadTimeout: time.Second * 5}
		res := ser.Serve(listener)
		log.Fatal("RPC HTTP server stopped", zap.String("result", res.Error()))
	}()
	return listener, nil
}

func WriteRPCResponseHTTP(w http.ResponseWriter, res gtypes.RPCResponse) {
	// jsonBytes := wire.JSONBytesPretty(res)
	jsonBytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	rww, ok := w.(*ResponseWriterWrapper)
	if ok {
		if res.Error != "" {
			rww.recordErr(errors.New(res.Error))
		} else {
			if res.Result != nil {
				rww.recordResponse(*res.Result)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonBytes)
}

//-----------------------------------------------------------------------------

// Wraps an HTTP handler, adding error logging.
// If the inner function panics, the outer function recovers, logs, sends an
// HTTP 500 error response.
func RecoverAndLogHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to remember the status
		traceId := r.Header.Get(utils.RequestIdHeader)
		begin := time.Now()
		if traceId == "" {
			traceId = utils.NewTraceId(begin).String()
			r.Header.Set(utils.RequestIdHeader, traceId)
		}
		query := r.URL.Query().Encode()
		path := r.URL.Path
		if query != "" {
			path += "?" + query
		}
		log.Audit().Info("rpc got request", zap.String("trace_id", traceId),
			zap.String("client_ip", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("path", path),
			zap.String("length", r.Header.Get("Content-Length")))
		rww := &ResponseWriterWrapper{
			ResponseWriter: w,
			buf:            bytes.NewBuffer(nil),
		}

		// Common headers
		origin := r.Header.Get("Origin")
		rww.Header().Set("Access-Control-Allow-Origin", origin)
		rww.Header().Set("Access-Control-Allow-Credentials", "true")
		rww.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
		rww.Header().Set("X-Server-Time", fmt.Sprintf("%v", begin.Unix()))

		defer func() {
			// Send a 500 error if a panic happens during a handler.
			// Without this, Chrome & Firefox were retrying aborted ajax requests,
			// at least to my localhost.
			if e := recover(); e != nil {

				// If RPCResponse
				if res, ok := e.(gtypes.RPCResponse); ok {
					WriteRPCResponseHTTP(rww, res)
				} else {
					// For the rest,
					log.Errorw("Panic in RPC HTTP handler", "error", e, "stack", string(debug.Stack()))
					fmt.Println(string(debug.Stack()))
					rww.WriteHeader(http.StatusInternalServerError)
					WriteRPCResponseHTTP(rww, gtypes.NewRPCResponse("", nil, gcmn.Fmt("Internal Server Error: %v", e)))
				}
			}

			// Finally, log.
			// durationMS := time.Since(begin).Nanoseconds() / 1000000
			// if rww.Status == -1 {
			// 	rww.Status = 200
			// }
			// log.Debug("Served RPC HTTP response",
			// 	zap.String("method", r.Method), zap.Stringer("url", r.URL),
			// 	zap.Int("status", rww.Status), zap.Int64("duration", durationMS),
			// 	zap.String("remoteAddr", r.RemoteAddr))
		}()
		handler.ServeHTTP(rww, r)
		reqDuration := time.Since(begin)
		var fields []zapcore.Field
		fields = append(fields, zap.String("trace_id", traceId),
			zap.Int("status", rww.Status),
			zap.Stringer("req_duration", reqDuration),
			zap.Int("length", len(rww.Data())),
			zap.Error(rww.err),
		)
		if len(rww.jsonRpcMethod) > 0 {
			fields = append(fields, zap.String("json_rpc_method", rww.jsonRpcMethod))
		}
		if len(rww.requestContent) > 0 {
			fields = append(fields, zap.ByteString("request_content", rww.requestContent))
		}
		if len(rww.responseContent) > 0 {
			fields = append(fields, zap.ByteString("response_content", rww.responseContent))
		}
		for k, v := range rww.logFields {
			fields = append(fields, zap.String(k, v))
		}
		log.Audit().Info("rpc got response ", fields...)
		rww.Flush()
	})
}

// Remember the status for logging
type ResponseWriterWrapper struct {
	Status int
	http.ResponseWriter
	buf             *bytes.Buffer
	jsonRpcMethod   string
	err             error
	requestContent  []byte
	responseContent []byte
	logFields       map[string]string
}

func (w *ResponseWriterWrapper) WriteHeader(status int) {
	if status == 0 {
		status = 200
	}
	w.Status = status
}

// implements http.Hijacker
func (w *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *ResponseWriterWrapper) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *ResponseWriterWrapper) Data() (b []byte) {
	return w.buf.Bytes()
}

func (w *ResponseWriterWrapper) recordErr(err error) {
	w.err = err
}

func (w *ResponseWriterWrapper) Flush() {
	w.ResponseWriter.WriteHeader(w.Status)
	w.ResponseWriter.Write(w.buf.Bytes())
	w.buf.Reset()
	w.err = nil
	w.requestContent = nil
	w.responseContent = nil
}

func (w *ResponseWriterWrapper) recordJsonRpcMethod(jsonRpcMethod string) {
	w.jsonRpcMethod = jsonRpcMethod
}

func (w *ResponseWriterWrapper) recordResponse(data []byte) {
	if len(data) > MaxAuditLogContentSize {
		w.responseContent = data[:MaxAuditLogContentSize]
		return
	}
	w.responseContent = data
}

func (w *ResponseWriterWrapper) recordRequest(data []byte) {
	if len(data) > MaxAuditLogContentSize {
		w.requestContent = data[:MaxAuditLogContentSize]
		return
	}
	w.requestContent = data
}

func (w *ResponseWriterWrapper) SetLogFields(fields map[string]string) {
	w.logFields = fields
}
