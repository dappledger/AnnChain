package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	grpc2 "github.com/dappledger/AnnChain/chain/proto"
	"github.com/dappledger/AnnChain/gemmill/config"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	rpcserver "github.com/dappledger/AnnChain/gemmill/rpc/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func runNode(testDir string) (node *Node, err error) {
	conf := config.DefaultConfig()
	conf.Set("environment", "development")
	conf.Set("log_path", "output.log")
	conf.Set("audit_log_path", "audit.log")
	err = config.InitRuntime(testDir, "test123", conf)
	if err != nil {
		return
	}
	node, err = NewNode(conf, testDir, conf.GetString("app_name"))
	if err != nil {
		return nil, fmt.Errorf("failed to new node: %v", err)
	}
	if err := node.Start(); err != nil {
		return nil, fmt.Errorf("failed to start node: %v", err)
	}
	if conf.GetString("rpc_laddr") != "" {
		if _, err := node.StartRPC(); err != nil {
			return nil, fmt.Errorf("failed to start rpc: %v", err)
		}
	}
	if conf.GetString("grpc_laddr") != "" {
		if err := node.StartGRPC(); err != nil {
			return nil, fmt.Errorf("failed to start rpc: %v", err)
		}
	}
	return
}

func stopNode(node *Node, testDir string) {
	if node != nil {
		node.Stop()
		time.Sleep(time.Second * 2)
	}
	err := os.RemoveAll(testDir)
	if err != nil {
		log.Errorf("remove test file failed %v", err)
	}
}

func TestGrpc(t *testing.T) {
	var ip = "127.0.0.1"
	var grpcAddr = ip + ":20981"
	var grpcGateway = "http://" + ip + ":20980"
	var testDir = "grpcTestData755756"
	node, err := runNode(testDir)
	defer stopNode(node, testDir)
	assert.NoError(t, err)
	time.Sleep(time.Second * 3)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rpcserver.MaxAuditLogContentSize = 4096
	cc, err := grpc.DialContext(ctx, grpcAddr, grpc.WithBlock(), grpc.WithInsecure())
	assert.NoError(t, err)
	defer cc.Close()
	client := grpc2.NewRpcServiceClient(cc)
	empty := grpc2.EmptyRequest{}
	respHeight, err := client.LastHeight(ctx, &empty)
	assert.NoError(t, err)
	log.Infof("grpc got  response %v", respHeight)
	empty = grpc2.EmptyRequest{}
	respInfo, err := client.NetInfo(ctx, &empty)
	assert.NoError(t, err)
	log.Infof("grpc got response %v", respInfo)
	req, err := http.NewRequest("GET", grpcGateway+"/api/v1/net_info", nil)
	assert.NoError(t, err)
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200, resp.Status)
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	log.Infof("http got response %v", string(data))
	req, err = http.NewRequest("GET", grpcGateway+"/api/v1/last_height", nil)
	assert.NoError(t, err)
	resp, err = httpClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200, resp.Status)
	data, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	log.Infof("http got response %v", string(data))

}
