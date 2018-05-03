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
