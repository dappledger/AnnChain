package kubernets

import (
	"bytes"
	"fmt"
	"github.com/dappledger/AnnChain/cmd/client/utils/io"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
	"github.com/dappledger/AnnChain/gemmill/go-wire"
	"github.com/dappledger/AnnChain/gemmill/types"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

type genesisBlock struct {
	Name string
	Helm HelmChart
}

type validatorSvc struct {
	Name      string
	Selector  string
	Namespace string
	Helm      HelmChart
}

func (v *validatorSvc) ValidatorRPC() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", v.Name, v.Namespace)
}

type validator struct {
	Name          string
	NamePrefix    string
	ImageName     string
	ImageTag      string
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
	Helm          HelmChart
}

func setupValidators(opt *Option) ([]validator, error) {

	validators := make([]validator, 0, opt.ValidatorNum)

	genDoc := &types.GenesisDoc{
		ChainID:    opt.ChainID,
		Plugins:    "specialop,querycache",
		Validators: make([]types.GenesisValidator, 0, opt.ValidatorNum),
	}
	crypto.NodeInit(opt.CryptoType)

	for i := 0; i < opt.ValidatorNum; i++ {
		privVal, err := types.GenPrivValidator(opt.CryptoType, nil)
		if err != nil {
			return nil, err
		}

		jsonBytes := wire.JSONBytes(privVal)
		privkey := privVal.PrivKey
		pubkey := privkey.PubKey()

		addr := pubkey.Address()

		v := validator{
			Name:          fmt.Sprintf("%s-validator%x-%d", opt.NamePrefix, addr[:4], i),
			NamePrefix:    fmt.Sprintf("%s-validator", opt.NamePrefix),
			ImageName:     opt.Rendez.Name,
			ImageTag:      opt.Rendez.Tag,
			ChainID:       opt.ChainID,
			P2PPort:       fmt.Sprintf("%d", 46000),
			RPCPort:       fmt.Sprintf("%d", 47000),
			SignByCA:      "false",
			PrivValidator: string(jsonBytes),
			Address:       fmt.Sprintf("%X", addr),
			PrivKey:       fmt.Sprintf("%x", privkey.Bytes()),
			PubKey:        fmt.Sprintf("%x", pubkey.Bytes()),
		}
		validators = append(validators, v)

		genDoc.Validators = append(genDoc.Validators, types.GenesisValidator{
			PubKey: privVal.PubKey,
			Amount: 100,
			IsCA:   true,
		})
	}

	seeds := make([]string, 0, len(validators))
	for _, v := range validators {
		seeds = append(seeds, fmt.Sprintf("%s.%s.svc.cluster.local:%s", v.Name, opt.Namespace, v.P2PPort))
	}
	seedsJoined := strings.Join(seeds, ",")

	genesisDoc := wire.JSONBytesPretty(genDoc)
	for i, _ := range validators {
		validators[i].GenesisDoc = string(genesisDoc)
		validators[i].Seeds = seedsJoined
	}

	src := filepath.Join(opt.OutputDir, "charts", "validator")
	bs, err := ioutil.ReadFile(filepath.Join(src, "values.yaml"))
	if err != nil {
		return nil, err
	}
	validatorValTpl, err := template.New("validator").Parse(string(bs))
	if err != nil {
		return nil, err
	}

	for i, v := range validators {
		relPath := filepath.Join("charts", fmt.Sprintf("validator-%d", i))
		dst := filepath.Join(opt.OutputDir, relPath)
		if err := io.Copy(src, dst); err != nil {
			return nil, err
		}
		buf := bytes.Buffer{}
		if err := validatorValTpl.Execute(&buf, v); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(filepath.Join(dst, "values.yaml"), buf.Bytes(), 0644); err != nil {
			return nil, err
		}

		validators[i].Helm = HelmChart{
			Name:         v.Name,
			ChartPath:    dst,
			RelChartPath: relPath,
		}
	}

	return validators, nil
}

func setupValidatorSvc(opt Option) (*validatorSvc, error) {
	valSvc := validatorSvc{
		Name:      opt.NamePrefix + "-validator-svc",
		Selector:  opt.NamePrefix + "-validator",
		Namespace: opt.Namespace,
		Helm: HelmChart{
			Name:         opt.NamePrefix + "-validator-svc",
			ChartPath:    filepath.Join(opt.OutputDir, "charts", "validator-svc"),
			RelChartPath: filepath.Join("charts", "validator-svc"),
		},
	}

	src := filepath.Join(opt.OutputDir, "charts", "validator-svc")
	bs, err := ioutil.ReadFile(filepath.Join(src, "values.yaml"))
	if err != nil {
		return nil, err
	}
	svcTpl, err := template.New("validator-svc").Parse(string(bs))
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	if err := svcTpl.Execute(&buf, valSvc); err != nil {
		return nil, err
	}
	dst := filepath.Join(opt.OutputDir, "charts", "validator-svc", "values.yaml")
	if err := ioutil.WriteFile(dst, buf.Bytes(), 0644); err != nil {
		return nil, err
	}
	return &valSvc, nil
}

func setupGenesisBlock(opt Option, genesisDoc []byte) (*genesisBlock, error) {

	genesisFile := filepath.Join(opt.OutputDir, "charts", "genesis-block", "genesis.json")
	if err := ioutil.WriteFile(genesisFile, genesisDoc, 0644); err != nil {
		return nil, err
	}
	gb := &genesisBlock{
		Name: opt.NamePrefix + "-genesis-block",
		Helm: HelmChart{
			Name:         opt.NamePrefix + "-genesis-block",
			ChartPath:    filepath.Join(opt.OutputDir, "charts", "genesis-block"),
			RelChartPath: filepath.Join("charts", "genesis-block"),
		},
	}
	return gb, nil
}
