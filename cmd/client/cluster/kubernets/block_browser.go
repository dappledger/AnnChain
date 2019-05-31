package kubernets

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

type blockBrowser struct {
	BindPort     string
	ImageName    string
	ImageTag     string
	IP           string
	MgoDB        string
	ValidatorRPC string
	Helm         HelmChart
}

func setupBlockBrowser(opt Option, validatorRPC string) (*blockBrowser, error) {

	browser := blockBrowser{
		BindPort:     "9090",
		ImageName:    opt.BlockBrowser.Name,
		ImageTag:     opt.BlockBrowser.Tag,
		ValidatorRPC: validatorRPC,
	}

	fname := filepath.Join(opt.OutputDir, "charts", "block-browser", "values.yaml")
	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	browserTpl, err := template.New("block-browser").Parse(string(bs))
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if err := browserTpl.Execute(&buf, browser); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(fname, buf.Bytes(), 0644); err != nil {
		return nil, err
	}
	browser.Helm = HelmChart{
		Name:         opt.NamePrefix + "-block-browser",
		ChartPath:    filepath.Join(opt.OutputDir, "charts", "block-browser"),
		RelChartPath: filepath.Join("charts", "block-browser"),
	}
	return &browser, nil
}
