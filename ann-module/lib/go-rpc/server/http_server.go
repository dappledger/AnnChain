// Commons for HTTP handling
package rpcserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	//	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"

	//. "github.com/dappledger/AnnChain/ann-module/lib/go-common"
	. "github.com/dappledger/AnnChain/ann-module/lib/go-rpc/types"
	//"github.com/dappledger/AnnChain/ann-module/lib/go-wire"
)

func StartHTTPServer(logger *zap.Logger, listenAddr string, handler http.Handler) (listener net.Listener, err error) {
	// listenAddr should be fully formed including tcp:// or unix:// prefix
	var proto, addr string
	parts := strings.SplitN(listenAddr, "://", 2)
	if len(parts) != 2 {
		logger.Warn("WARNING (go-rpc): Please use fully formed listening addresses, including the tcp:// or unix:// prefix")
		// we used to allow addrs without tcp/unix prefix by checking for a colon
		// TODO: Deprecate
		proto = SocketType(listenAddr)
		addr = listenAddr
		// return nil, fmt.Errorf("Invalid listener address %s", lisenAddr)
	} else {
		proto, addr = parts[0], parts[1]
	}

	logger.Sugar().Infof("Starting RPC HTTP server on %s socket %v", proto, addr)
	listener, err = net.Listen(proto, addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to listen to %v: %v", listenAddr, err)
	}

	go func() {
		// res := http.Serve(
		// 	listener,
		// 	RecoverAndLogHandler(handler),
		// )
		ser := &http.Server{Handler: RecoverAndLogHandler(logger, handler), ReadTimeout: time.Second * 5}
		res := ser.Serve(listener)
		logger.Fatal("RPC HTTP server stopped", zap.String("result", res.Error()))
	}()
	return listener, nil
}

func WriteRPCResponseHTTP(w http.ResponseWriter, res RPCResponse) {
	// jsonBytes := wire.JSONBytesPretty(res)
	jsonBytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(200)
	w.Write(jsonBytes)
}

//-----------------------------------------------------------------------------

// Wraps an HTTP handler, adding error logging.
// If the inner function panics, the outer function recovers, logs, sends an
// HTTP 500 error response.
func RecoverAndLogHandler(logger *zap.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to remember the status
		rww := &ResponseWriterWrapper{-1, w}
		begin := time.Now()

		// Common headers
		origin := r.Header.Get("Origin")
		rww.Header().Set("Access-Control-Allow-Origin", origin)
		rww.Header().Set("Access-Control-Allow-Credentials", "true")
		rww.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
		rww.Header().Set("X-Server-Time", fmt.Sprintf("%v", begin.Unix()))

		//		defer func() {
		//			// Send a 500 error if a panic happens during a handler.
		//			// Without this, Chrome & Firefox were retrying aborted ajax requests,
		//			// at least to my localhost.
		//			if e := recover(); e != nil {

		//				// If RPCResponse
		//				if res, ok := e.(RPCResponse); ok {
		//					WriteRPCResponseHTTP(rww, res)
		//				} else {
		//					// For the rest,
		//					logger.Sugar().Errorw("Panic in RPC HTTP handler", "error", e, "stack", string(debug.Stack()))
		//					rww.WriteHeader(http.StatusInternalServerError)
		//					WriteRPCResponseHTTP(rww, NewRPCResponse("", nil, -32000, Fmt("Internal Server Error: %v", e)))
		//				}
		//			}

		//			// Finally, log.
		//			durationMS := time.Since(begin).Nanoseconds() / 1000000
		//			if rww.Status == -1 {
		//				rww.Status = 200
		//			}
		//			logger.Debug("Served RPC HTTP response",
		//				zap.String("method", r.Method), zap.Stringer("url", r.URL),
		//				zap.Int("status", rww.Status), zap.Int64("duration", durationMS),
		//				zap.String("remoteAddr", r.RemoteAddr))
		//		}()

		handler.ServeHTTP(rww, r)
	})
}

// Remember the status for logging
type ResponseWriterWrapper struct {
	Status int
	http.ResponseWriter
}

func (w *ResponseWriterWrapper) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

// implements http.Hijacker
func (w *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
