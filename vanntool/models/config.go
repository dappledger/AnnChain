package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var config Config

type Config struct {
	AnntoolPath string
	DataPath    string
	JvmFilePath string
}

func (c *Config) init() {
	buf, err := ioutil.ReadFile("./conf/config.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(buf, c); err != nil {
		panic(err)
	}
	fmt.Println("read config:", config)
}

func (c *Config) GetAnntoolPath() string {
	if len(c.AnntoolPath) == 0 {
		return "./bin/anntool"
	}

	return os.ExpandEnv(c.AnntoolPath)
}

func (c *Config) GetDataPath() string {
	if len(c.DataPath) == 0 {
		c.DataPath = "./data/data.db"
	}

	c.DataPath = os.ExpandEnv(c.DataPath)
	if err := EnsureDir(c.DataPath); err != nil {
		panic(err)
	}

	return c.DataPath
}

func (c *Config) GetJvmFilePath() string {
	if len(c.JvmFilePath) == 0 {
		c.JvmFilePath = "./data/contractfiles/"
	}

	c.JvmFilePath = os.ExpandEnv(c.JvmFilePath)
	if err := EnsureDir(c.JvmFilePath); err != nil {
		panic(err)
	}

	return c.JvmFilePath
}
