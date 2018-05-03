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
)

type NodesOpController struct {
	beego.Controller
}

func (c *NodesOpController) Get() {
	c.Data["nodes"] = models.NodeM().ListNode()
	beego.Debug("[nods_ctr],list:", models.NodeM().ListNode())
	c.TplName = "nodesop.tpl"
}

const RESULT_OK = "ok"

func (c *NodesOpController) Post() {
	nodeName := c.Input().Get("nodename")
	nodeIP := c.Input().Get("nodeip")
	nodePk := c.Input().Get("encnodepk")
	method := c.Input().Get("method")
	res := RESULT_OK
	switch method {
	case "add":
		if nodeName == "" || nodeIP == "" || nodePk == "" {
			res = "请补全 节点名|节点IP|节点私钥"
		} else {
			err := models.NodeM().Insert(nodeName, nodeIP, nodePk)
			if err != nil {
				beego.Warn("[nodes_op],insert node error:", err)
				res = err.Error()
			}
		}
	case "modify":
		if nodeName == "" || nodeIP == "" || nodePk == "" {
			res = "请补全 节点名|节点IP|节点私钥"
		} else {
			err := models.NodeM().Modify(nodeName, nodeIP, nodePk)
			if err != nil {
				beego.Warn("[nodes_op],modify node error:", err)
				res = err.Error()
			}
		}
	case "delete":
		if nodeName == "" {
			res = "请补全 节点名"
		} else {
			err := models.NodeM().Drop(nodeName)
			if err != nil {
				beego.Warn("[nodes_op],delete node error:", err)
				res = err.Error()
			}
		}
	default:
		res = "方法缺失"
	}
	c.Data["json"] = res
	if res == RESULT_OK {
		c.Data["nodes"] = models.NodeM().ListNode()
	}
	c.ServeJSON()
}
