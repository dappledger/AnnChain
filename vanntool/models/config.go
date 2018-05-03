/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


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
