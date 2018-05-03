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
