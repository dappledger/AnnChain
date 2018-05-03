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
	"fmt"

	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/vanntool/models"
	"github.com/dappledger/AnnChain/vanntool/tools"
)

type CmdListController struct {
	beego.Controller
}

func (c *CmdListController) Get() {
	c.TplName = "cmdlist.tpl"
}

func (c *CmdListController) Post() {
	cmd := c.Input().Get("cmd")
	op := c.Input().Get("op")

	var res = fmt.Sprintf("hello world:%v,%v", cmd, op)
	if do := models.GetCmdOp(cmd, op); do != nil {
		if err := do.FillData(&c.Controller); err != nil {
			beego.Warn("[cmd_list],cmd op fill data failed,cmd:", cmd, ",op:", op, ",err:", err)
			res = err.Error()
		} else {
			res = tools.ParseRet(do.Do())
		}
	}
	c.Data["json"] = &res
	c.ServeJSON()
}

type CmdListPastController struct {
	beego.Controller
}

func (c *CmdListPastController) Get() {
	c.TplName = "cmdlist-past.tpl"
}

func (c *CmdListPastController) Post() {
	cmd := c.Input().Get("cmd")
	op := c.Input().Get("op")

	var res = fmt.Sprintf("hello world:%v,%v", cmd, op)
	if do := models.GetCmdOp(cmd, op); do != nil {
		if err := do.FillData(&c.Controller); err != nil {
			beego.Warn("[cmd_list],cmd op fill data failed,cmd:", cmd, ",op:", op, ",err:", err)
			res = err.Error()
		} else {
			res = tools.ParseRet(do.Do())
		}
	}
	c.Data["json"] = &res
	c.ServeJSON()
}
