package config

import (
	"testing"
	"fmt"
)

var testCfgFile = "./cfg.json"

func TestParse(t *testing.T) {
	cfg := LoadConfigFile(testCfgFile)
	if cfg.GetFloat("port") != 8010 || cfg.GetString("role") != "cc" || cfg.GetBool("idgen") == false || cfg.GetFloat("sid") != 0 {
		t.Error("Fatal")
	}

	// fmt.Println(cfg.GetFloat("port"))
	fmt.Println(cfg.GetInt("port"))
}
