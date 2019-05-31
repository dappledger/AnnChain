package kubernets

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dappledger/AnnChain/cmd/client/cluster"
	"github.com/dappledger/AnnChain/gemmill/go-crypto"
)

//go:generate directory2go -dst chart.go -pkg kubernets -src charts

type Helm struct {
	opt           Option
	validators    []validator
	validatorsSvc *validatorSvc
	browser       *blockBrowser
	genesisBlock  *genesisBlock
}

type HelmChart struct {
	Name         string
	Args         []string
	ChartPath    string
	RelChartPath string
}

func (h *HelmChart) Install() error {

	cmd := exec.Command("helm", "install")
	if h.Name != "" {
		cmd.Args = append(cmd.Args, "--name", h.Name)
	}

	if len(h.Args) > 0 {
		cmd.Args = append(cmd.Args, "--set", strings.Join(h.Args, ","))
	}

	cmd.Args = append(cmd.Args, h.ChartPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	fmt.Println(string(output))
	return nil
}

func (h *HelmChart) InstallScript() string {

	args := make([]string, 0, 10)
	args = append(args, "helm", "install")
	if h.Name != "" {
		args = append(args, "--name", h.Name)
	}

	if len(h.Args) > 0 {
		args = append(args, "--set", strings.Join(h.Args, ","))
	}
	args = append(args, h.RelChartPath)
	return strings.Join(args, " ")
}

func (h *HelmChart) Delete() error {
	cmd := exec.Command("helm", "delete", h.Name, "--purge")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

func (h *HelmChart) DeleteScript() string {

	args := []string{"helm", "delete", h.Name, "--purge"}
	return strings.Join(args, " ")
}

type Image struct {
	Name string
	Tag  string
}

type Option struct {
	ValidatorNum int
	HasBrowser   bool
	HasAPI       bool
	NamePrefix   string
	ChainID      string
	CryptoType   string
	Namespace    string // for k8s
	OutputDir    string
	Rendez       *Image
	BlockBrowser *Image
}

var defaultOption = Option{
	ValidatorNum: 1,
	OutputDir:    filepath.Join(os.TempDir(), fmt.Sprintf("ann-helm-%d", time.Now().Unix())),
	Namespace:    "default",
	CryptoType:   crypto.CryptoTypeZhongAn,
	Rendez:       &Image{"rendez", "latest"},
	BlockBrowser: &Image{"block-browser", "latest"},
}

func checkOption(opt Option) (option Option, err error) {

	if opt == (Option{}) {
		return defaultOption, nil
	}

	if opt.ValidatorNum == 0 {
		opt.ValidatorNum = 1
	}

	if opt.OutputDir == "" {
		opt.OutputDir = filepath.Join(os.TempDir(), fmt.Sprintf("ann-helm-%d", time.Now().Unix()))
	}

	if opt.CryptoType == "" {
		opt.CryptoType = crypto.CryptoTypeZhongAn
	}

	if opt.Namespace == "" {
		opt.Namespace = "default"
	}

	if opt.Rendez == nil {
		opt.Rendez = &Image{"rendez", "latest"}
	}

	if opt.BlockBrowser == nil {
		opt.BlockBrowser = &Image{"block-browser", "latest"}
	}
	return opt, nil
}

func NewHelm(opt Option) (*Helm, error) {

	opt, err := checkOption(opt)
	if err != nil {
		return nil, err
	}

	h := Helm{opt: opt}
	if err := GenerateFiles(h.opt.OutputDir); err != nil {
		return nil, fmt.Errorf("generate charts template err %v", err)
	}

	validators, err := setupValidators(&opt)
	if err != nil {
		return nil, err
	}
	h.validators = validators

	svc, err := setupValidatorSvc(opt)
	if err != nil {
		return nil, err
	}
	h.validatorsSvc = svc

	if opt.HasBrowser {
		browser, err := setupBlockBrowser(opt, h.validatorsSvc.ValidatorRPC())
		if err != nil {
			return nil, err
		}
		h.browser = browser
	}

	gb, err := setupGenesisBlock(opt, []byte(h.validators[0].GenesisDoc))
	if err != nil {
		return nil, err
	}
	h.genesisBlock = gb

	h.SaveScripts()
	return &h, nil
}

func (h *Helm) Up() error {

	if err := h.genesisBlock.Helm.Install(); err != nil {
		return err
	}

	for _, v := range h.validators {
		if err := v.Helm.Install(); err != nil {
			return err
		}
	}
	if err := h.validatorsSvc.Helm.Install(); err != nil {
		return err
	}

	if h.opt.HasBrowser {
		if err := h.browser.Helm.Install(); err != nil {
			return err
		}
	}
	return nil
}

func (h *Helm) Down() error {

	if err := h.genesisBlock.Helm.Delete(); err != nil {
		return err
	}
	if err := h.validatorsSvc.Helm.Delete(); err != nil {
		return err
	}

	for _, v := range h.validators {
		if err := v.Helm.Delete(); err != nil {
			return err
		}
	}

	if h.opt.HasBrowser {
		if err := h.browser.Helm.Delete(); err != nil {
			return err
		}
	}
	return nil
}

func (h *Helm) GetValidatorInfo(index int) (*cluster.ValidatorInfo, error) {

	v := h.validators[index]

	info := &cluster.ValidatorInfo{
		RPC:     fmt.Sprintf("%s.%s.svc.cluster.local:%s", v.Name, h.opt.Namespace, v.RPCPort),
		Address: v.Address,
	}
	return info, nil
}

func (h *Helm) StartValidator(index int) error {
	panic("implement me")
}

func (h *Helm) StopValidator(index int) error {
	panic("implement me")
}

func (h *Helm) SaveScripts() error {

	scriptBuf := &bytes.Buffer{}
	scriptBuf.WriteString("#!/bin/bash\n")

	scriptBuf.WriteString(h.genesisBlock.Helm.InstallScript())
	scriptBuf.WriteString("\n")
	for _, v := range h.validators {
		scriptBuf.WriteString(v.Helm.InstallScript())
		scriptBuf.WriteString("\n")
	}
	scriptBuf.WriteString(h.validatorsSvc.Helm.InstallScript())
	scriptBuf.WriteString("\n")

	if h.opt.HasBrowser {
		scriptBuf.WriteString(h.browser.Helm.InstallScript())
		scriptBuf.WriteString("\n")
	}

	installScriptPath := filepath.Join(h.opt.OutputDir, "install.sh")
	if err := ioutil.WriteFile(installScriptPath, scriptBuf.Bytes(), 0777); err != nil {
		return err
	}

	scriptBuf.Reset()
	scriptBuf.WriteString("#!/bin/bash\n")

	scriptBuf.WriteString(h.genesisBlock.Helm.DeleteScript())
	scriptBuf.WriteString("\n")
	for _, v := range h.validators {
		scriptBuf.WriteString(v.Helm.DeleteScript())
		scriptBuf.WriteString("\n")
	}
	scriptBuf.WriteString(h.validatorsSvc.Helm.DeleteScript())
	scriptBuf.WriteString("\n")

	if h.opt.HasBrowser {
		scriptBuf.WriteString(h.browser.Helm.DeleteScript())
		scriptBuf.WriteString("\n")
	}

	uninstallScriptPath := filepath.Join(h.opt.OutputDir, "delete.sh")
	if err := ioutil.WriteFile(uninstallScriptPath, scriptBuf.Bytes(), 0777); err != nil {
		return err
	}
	return nil
}
