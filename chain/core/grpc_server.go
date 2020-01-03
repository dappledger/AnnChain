package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	grpc2 "github.com/dappledger/AnnChain/chain/proto"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	server "github.com/dappledger/AnnChain/gemmill/rpc/server"
	gtypes "github.com/dappledger/AnnChain/gemmill/rpc/types"
	"github.com/gogo/gateway"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
)

type grpcHandler struct {
	node *Node
}

func newGRPCHandler(node *Node) *grpcHandler {
	return &grpcHandler{node: node}
}

var _ grpc2.RpcServiceServer = (*grpcHandler)(nil)

func (n *Node) startGrpc(listenAddr string) (proto, addr string, err error) {
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_validator.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(log.Audit()),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log.Audit()),
			grpc_recovery.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
		)),
	)
	grpc2.RegisterRpcServiceServer(s, newGRPCHandler(n))

	// Serve gRPC Server
	parts := strings.SplitN(listenAddr, "://", 2)
	if len(parts) != 2 {
		proto = gtypes.SocketType(listenAddr)
		addr = listenAddr
	} else {
		proto, addr = parts[0], parts[1]
	}
	log.Infof("Starting gRPC server on [%s://%s ", proto, addr)
	listener, err := net.Listen(proto, addr)
	if err != nil {
		err = fmt.Errorf("failed to listen to %v: %v", listenAddr, err)
		return
	}
	go func() {
		err = s.Serve(listener)
		log.Fatal("gRPC stopped", zap.Error(err))
	}()
	return
}

func (n *Node) startGrpcGateway(grpcAddr, proto string, gatewayPort int) error {
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	jsonpb := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
		// This is necessary to get error details properly
		// marshalled in unary requests.
		runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),
	)
	ctx := context.Background()
	conn, err := dial(ctx, proto, grpcAddr)
	if err != nil {
		return fmt.Errorf("grpc gateway dial err %v", err)
	}
	err = grpc2.RegisterRpcServiceHandler(ctx, gwmux, conn)
	if err != nil {
		return fmt.Errorf("failed to register gateway:%v", err)
	}
	ser := &http.Server{
		Handler:     server.RecoverAndLogHandler(gwmux),
		ReadTimeout: time.Second * 5,
		Addr:        ":" + fmt.Sprintf("%d", gatewayPort),
	}
	go func() {
		<-ctx.Done()
		log.Infof("Shutting down the http server")
		if err := ser.Shutdown(context.Background()); err != nil {
			log.Warnf("failed to shutdown http server: %v", err)
		}
	}()

	go func() {
		err = ser.ListenAndServe()
		log.Fatal("gRPC HTTP server stopped", zap.Error(err))
	}()
	return nil
}

func dial(ctx context.Context, network, addr string) (*grpc.ClientConn, error) {
	switch network {
	case "tcp":
		parts := strings.SplitN(addr, ":", 2)
		if len(parts) == 2 && (parts[0] == "0.0.0.0" || parts[0] == "") {
			addr = "localhost:" + parts[1]
		}
		return dialTCP(ctx, addr)
	case "unix":
		return dialUnix(ctx, addr)
	default:
		return nil, fmt.Errorf("unsupported network type %q", network)
	}
}

// dialTCP creates a client connection via TCP.
// "addr" must be a valid TCP address with a port number.
func dialTCP(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, grpc.WithInsecure())
}

// dialUnix creates a client connection via a unix domain socket.
// "addr" must be a valid path to the socket.
func dialUnix(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	d := func(addr string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", addr, timeout)
	}
	return grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithDialer(d))
}
