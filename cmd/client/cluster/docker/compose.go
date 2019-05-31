package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dappledger/AnnChain/cmd/client/cluster"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/types"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	composeTplStr = `version: '3'
services:
  {{- range .Services }}
  {{ . }}
  {{- end }}
networks:
  app_net:
    driver: bridge
    ipam:
      driver: default
      config:
      - subnet: {{ .IPPrefix }}.0/24`

	mongoDBTplStr = `mgo:
    image: {{.Image}}
    ports:
      - '{{.BindPort}}:{{.BindPort}}'
    command: mongod
    restart: always
    networks:
      app_net:
        ipv4_address: {{.IP}}`

	blockBrowserServiceTplStr = `block-browser:
    image: {{.Image}}
    ports:
      - '{{.BindPort}}:{{.BindPort}}'
    entrypoint:
      - /bin/sh
      - -c
      - |
        echo 'appname = hubble
              httpport = {{.BindPort}}
              runmode ="dev"

              # blockchain's api ,not app's api
              api_addr = "{{.ValidatorRPC}}"
              chain_id = "1024"

              mogo_addr = "{{.MgoDB}}"
              #mogo_user = hzc
              #mogo_pwd = 123456

              mogo_db = "ann-browser"
              sync_job = 1
              sync_from_height = 0' > conf/app.conf
        block-browser
    restart: always
    networks:
      app_net:
        ipv4_address: {{.IP}}`

	validatorServiceTplStr = `{{.Name}}:
    hostname: {{.HostName}}
    image: {{.Image}}
    ports:
      - '{{.P2PPort}}:{{.P2PPort}}'
      - '{{.RPCPort}}:{{.RPCPort}}'
    entrypoint:
      - /bin/sh
      - -c
      - |
        mkdir -p logs
        mkdir -p /rendez
        rendez init --runtime="/rendez" --log_path="logs/rendez.log" --app="evm" --chainid={{.ChainID}}
        echo '{{.GenesisDoc}}' > /rendez/genesis.json
        echo 'app_name = "evm"
              auth_by_ca = false
              block_size = 5000
              crypto_type = "{{.CryptoType}}"
              db_backend = "leveldb"
              db_conn_str = "sqlite3"
              db_type = "sqlite3"
              environment = "production"
              fast_sync = true
              log_path = ""
              moniker = "anonymous"
              non_validator_auth_by_ca = false
              non_validator_node_auth = false
              p2p_laddr = "tcp://0.0.0.0:{{.P2PPort}}"
              rpc_laddr = "tcp://0.0.0.0:{{.RPCPort}}"
              seeds = "{{.Seeds}}"
              signbyca = "{{.SignByCA}}"
              skip_upnp = true
              threshold_blocks = 0
              tracerouter_msg_ttl = 5
              network_rate_limit = 1024' > /rendez/config.toml
        echo '{{.PrivValidator}}' > /rendez/priv_validator.json
        rendez run --runtime="/rendez"
    networks:
      app_net:
        ipv4_address: {{.IP}}
    restart: always`
)

type mgo struct {
	BindPort string
	Image    string
	IP       string
	dockerTypes.Container
}

type blockBrowser struct {
	BindPort     string
	Image        string
	IP           string
	MgoDB        string
	ValidatorRPC string
	dockerTypes.Container
}

type rendezAPI struct {
	BindPort     string
	Image        string
	ValidatorRPC string
	IP           string
	dockerTypes.Container
}

type validator struct {
	Name          string
	Image         string
	HostName      string
	ChainID       string
	P2PPort       string
	RPCPort       string
	Seeds         string
	SignByCA      string
	Address       string
	PrivValidator string
	GenesisDoc    string
	PrivKey       string
	PubKey        string
	IP            string
	CryptoType    string
	dockerTypes.Container
}

type DockerCompose struct {
	Browser    *blockBrowser
	API        *rendezAPI
	DB         *mgo
	Validators []validator
	runtimeDir string
	Opt        *Option
}

type Option struct {
	ValidatorNum      int
	IPPrefix          string
	HasBrowser        bool
	HasAPI            bool
	RendezImage       string
	BlockBrowserImage string
	CryptoType        string
}

var DefaultOpt = Option{
	ValidatorNum:      1,
	IPPrefix:          "192.168.32",
	HasBrowser:        true,
	HasAPI:            true,
	RendezImage:       "rendez:latest",
	BlockBrowserImage: "block-browser:latest",
	CryptoType:        crypto.CryptoTypeZhongAn,
}

func NewDockerCompose(opt *Option) *DockerCompose {

	if opt == nil {
		opt = &DefaultOpt
	}

	validators := SetupValidators(opt.ValidatorNum, opt.CryptoType, opt.IPPrefix, opt.RendezImage)

	compose := &DockerCompose{
		Validators: validators,
		runtimeDir: filepath.Join(os.TempDir(), fmt.Sprintf("ann-docker-compose-%d", time.Now().Unix())),
		Opt:        opt,
	}

	if compose.Opt.HasBrowser {

		m := mgo{
			BindPort: "27017",
			Image:    "mongo:3.2",
			IP:       compose.Opt.IPPrefix + ".7",
		}
		browser := blockBrowser{
			BindPort:     "9090",
			MgoDB:        m.IP + ":" + m.BindPort,
			Image:        opt.BlockBrowserImage,
			IP:           compose.Opt.IPPrefix + ".8",
			ValidatorRPC: validators[0].HostName + ":" + validators[0].RPCPort,
		}
		compose.DB = &m
		compose.Browser = &browser
	}
	return compose
}

func (d *DockerCompose) PrintInfo() {

	if d.Opt.HasBrowser {
		fmt.Printf("block-browser: http://127.0.0.1:%s\tcontainerId:%s\n", d.Browser.BindPort, d.Browser.Container.ID)
		fmt.Printf("mongoDB: http://127.0.0.1:%s\tcontainerId:%s\n", d.DB.BindPort, d.DB.Container.ID)
	}
	if d.Opt.HasAPI {
		fmt.Printf("rendez-api: http://127.0.0.1:%s\tcontainerId:%s\n", d.API.BindPort, d.API.Container.ID)
	}
	fmt.Println("validators:")
	for _, v := range d.Validators {
		fmt.Printf("\t%s\tcontainerId:%s\n", v.Name, v.Container.ID)
	}
	fmt.Println("workDir:", d.runtimeDir)
}

func (d *DockerCompose) Up() error {

	if err := os.MkdirAll(d.runtimeDir, 0777); err != nil {
		return err
	}

	cfg, err := d.GenerateConfig()
	if err != nil {
		return err
	}

	ymlFile := filepath.Join(d.runtimeDir, "docker-compose.yml")
	if err := ioutil.WriteFile(ymlFile, []byte(cfg), 0666); err != nil {
		return err
	}

	cmd := exec.Command("docker-compose", "up")
	cmd.Dir = d.runtimeDir

	errChan := make(chan error, 1)
	go func() {
		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("%v, make sure previous docker-composes has been stopped, run 'docker-compose down'", err)
			errChan <- err
		}
	}()

	cli, err := client.NewClient(client.DefaultDockerHost, "1.39", nil, nil)
	if err != nil {
		return err
	}

	wait := func(cli *client.Client) (error, bool) {

		_, dockerNames := filepath.Split(d.runtimeDir)
		args := filters.NewArgs()
		args.Add("name", dockerNames)

		allContainers, err := cli.ContainerList(context.Background(), dockerTypes.ContainerListOptions{Filters: args})
		if err != nil {
			return err, false
		}

		expectContainersNum := len(d.Validators)

		if d.Opt.HasAPI {
			expectContainersNum += 1
		}

		if d.Opt.HasBrowser {
			expectContainersNum += 2
		}

		if len(allContainers) != expectContainersNum {
			return nil, false
		}

		for i, v := range allContainers {
			if !strings.HasPrefix(v.Status, "Up ") {
				return nil, false
			}
			if strings.Contains(v.Names[0], "_block-browser_") {
				d.Browser.Container = allContainers[i]
			} else if strings.Contains(v.Names[0], "_mgo_") {
				d.DB.Container = allContainers[i]
			} else if strings.Contains(v.Names[0], "_validator-") {
				u := strings.LastIndex(v.Names[0], "-")
				str := v.Names[0][u:]
				u = strings.LastIndex(str, "_")
				index, _ := strconv.Atoi(str[:u])
				d.Validators[index].Container = allContainers[i]
			}
		}
		return nil, true
	}

	for {
		select {
		case err := <-errChan:
			return err
		default:
			if err, done := wait(cli); err != nil {
				return err
			} else if done {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return nil
}

func (d *DockerCompose) Down() error {

	cmd := exec.Command("docker-compose", "down")
	cmd.Dir = d.runtimeDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return os.RemoveAll(d.runtimeDir)
}

func (d *DockerCompose) GetValidatorInfo(index int) (*cluster.ValidatorInfo, error) {

	return &cluster.ValidatorInfo{
		RPC:     "http://127.0.0.1:" + d.Validators[index].RPCPort,
		Address: d.Validators[index].Address,
	}, nil
}

func (d *DockerCompose) StopValidator(index int) error {

	if len(d.Validators) < index {
		return errors.New("invalid index")
	}

	cli, err := client.NewClient(client.DefaultDockerHost, "1.39", nil, nil)
	if err != nil {
		return err
	}
	timeout := time.Minute
	err = cli.ContainerStop(context.Background(), d.Validators[index].ID, &timeout)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerCompose) StartValidator(index int) error {

	if len(d.Validators) < index {
		return errors.New("invalid index")
	}

	cli, err := client.NewClient(client.DefaultDockerHost, "1.39", nil, nil)
	if err != nil {
		return err
	}

	err = cli.ContainerStart(context.Background(), d.Validators[index].ID, dockerTypes.ContainerStartOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerCompose) GenerateConfig() (string, error) {

	servicesTpl := make([]string, 0, len(d.Validators)+2)

	if d.Opt.HasBrowser {

		mgoTpl, err := executeTpl(d.DB, "mgo", mongoDBTplStr)
		if err != nil {
			return "", err
		}
		servicesTpl = append(servicesTpl, mgoTpl)

		browserTpl, err := executeTpl(d.Browser, "browser", blockBrowserServiceTplStr)
		if err != nil {
			return "", err
		}
		servicesTpl = append(servicesTpl, browserTpl)
	}

	for _, v := range d.Validators {
		tpl, err := executeTpl(v, "rendez", validatorServiceTplStr)
		if err != nil {
			return "", err
		}
		servicesTpl = append(servicesTpl, tpl)
	}

	compose := map[string]interface{}{
		"Services": servicesTpl,
		"IPPrefix": d.Opt.IPPrefix,
	}

	composeTpl, err := executeTpl(compose, "compose", composeTplStr)
	if err != nil {
		return "", err
	}
	return composeTpl, nil
}

func makeValidatorsIp(ipPrefix string, num int) []string {
	ips := make([]string, 0, num)
	for i := 0; i < num; i++ {
		ips = append(ips, ipPrefix+fmt.Sprintf(".%d", 10+i))
	}
	return ips
}

func makeValidatorsPorts(num int) (p2p []string, rpc []string) {

	for i := 0; i < num; i++ {
		p2p = append(p2p, fmt.Sprintf("%d", 46000+i))
		rpc = append(rpc, fmt.Sprintf("%d", 47000+i))
	}
	return
}

func makeSeeds(ips, ports []string) string {
	seeds := make([]string, 0, len(ips))

	for k, _ := range ips {
		seeds = append(seeds, ips[k]+":"+ports[k])
	}
	return strings.Join(seeds, ",")
}

func SetupValidators(num int, cryptoType string, ipPrefix string, image string) []validator {

	ips := makeValidatorsIp(ipPrefix, num)
	p2pPorts, rpcPorts := makeValidatorsPorts(num)
	seeds := makeSeeds(ips, p2pPorts)

	chainID := "9102"

	validators := make([]validator, 0, num)

	genDoc := &types.GenesisDoc{
		ChainID:    chainID,
		Plugins:    "specialop,querycache",
		Validators: make([]types.GenesisValidator, 0, num),
	}

	for i := 0; i < num; i++ {

		privVal, err := types.GenPrivValidator(cryptoType, nil)
		if err != nil {
			panic(err)
		}
		key := privVal.PrivKey
		jsonBytes := wire.JSONBytes(privVal)
		pubkey := key.PubKey()
		v := validator{
			Name:          fmt.Sprintf("validator-%d", i),
			Image:         image,
			HostName:      fmt.Sprintf("validator-%d", i),
			ChainID:       chainID,
			P2PPort:       p2pPorts[i],
			RPCPort:       rpcPorts[i],
			Seeds:         seeds,
			PrivValidator: string(jsonBytes),
			Address:       fmt.Sprintf("%X", pubkey.Address()),
			PrivKey:       fmt.Sprintf("%x", key.Bytes()),
			PubKey:        fmt.Sprintf("%x", pubkey.Bytes()),
			IP:            ips[i],
			CryptoType:    cryptoType,
		}
		validators = append(validators, v)
		genDoc.Validators = append(genDoc.Validators, types.GenesisValidator{
			PubKey: privVal.PubKey,
			Amount: 100,
			IsCA:   true,
		})
	}

	genesisDoc := wire.JSONBytes(genDoc)
	for i, _ := range validators {
		validators[i].GenesisDoc = string(genesisDoc)
	}
	return validators
}

func executeTpl(obj interface{}, tplName string, tplString string) (string, error) {

	tpl, err := template.New(tplName).Parse(tplString)
	if err != nil {
		return "", err
	}
	cfg := new(bytes.Buffer)
	if err := tpl.Execute(cfg, obj); err != nil {
		return "", err
	}
	return cfg.String(), nil
}
