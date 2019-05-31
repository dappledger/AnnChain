package kubernets

import (
	"fmt"
	"testing"
	"time"
)

func TestNewHelm(t *testing.T) {
	fmt.Println("he")
	opt := Option{
		ValidatorNum: 4,
		//OutputDir:    "/Users/zakj/vms/k8s/ann-helm",
		NamePrefix: "test",
		ChainID:    "9012",
		HasBrowser: true,
		HasAPI:     true,
	}
	h, err := NewHelm(opt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(h.opt.OutputDir)
	//return
	if err := h.Up(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Minute * 6)
	h.Down()
}
