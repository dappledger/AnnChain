package controllers

import (
	"github.com/astaxie/beego"
	"github.com/dappledger/AnnChain/src/vision/models"
)

type InitNode struct {
	beego.Controller

	Runtime string
}

func (c *InitNode) Get() {
	c.Data["runtime"] = c.Runtime
	c.TplName = "initnode.tpl"
}

func (c *InitNode) Post() {
	method := c.Input().Get("method")
	switch method {
	case "genkey":
		c.Data["json"] = models.GenKeyInfo()
	case "init":
		models.DoInitNode(&c.Controller, c.Runtime)
	case "run":
		models.RunNode(&c.Controller)
	case "close":
		models.CloseServer(&c.Controller)
	default:
		c.Data["json"] = "no default"
	}
	c.ServeJSON()
}
