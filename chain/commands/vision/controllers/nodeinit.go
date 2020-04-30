// Copyright © 2017 ZhongAn Technology
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"github.com/dappledger/AnnChain/chain/commands/vision/models"
)

func (c *InitNode) Get() {
	c.TplName = "initnode.tpl"
}

func (c *InitNode) Post() {
	method := c.Input().Get("method")
	switch method {
	case "genkey":
		var err error
		cryptoType := c.Input().Get("crypto_list")
		c.Data["json"], err = models.GenKeyInfo(cryptoType)
		if err != nil {
			c.Data["json"] = err.Error()
		}
	case "init":
		models.DoInitNode(&c.Controller)
	case "run":
		models.RunNode(&c.Controller)
	case "close":
		models.CloseServer(&c.Controller)

	default:
		c.Data["json"] = "no default"
	}
	c.ServeJSON()
}
