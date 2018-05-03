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


package controllers

import (
	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/vanntool/models"
	"github.com/dappledger/AnnChain/vanntool/tools"
)

type ShowNodeInfo struct {
	beego.Controller
}

func (c *ShowNodeInfo) Get() {
	c.TplName = "nodeinfo.tpl"
}

func (c *ShowNodeInfo) Post() {
	node := c.Input().Get("nodename")
	org := c.Input().Get("orgname")
	codeHash := c.Input().Get("code_hash")
	nodeShow, err := models.NodeM().ShowInfo(node, org, codeHash)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = tools.ParseRet(nodeShow.String())
	}
	c.ServeJSON()
}
