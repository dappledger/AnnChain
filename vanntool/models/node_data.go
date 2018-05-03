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

	"github.com/dappledger/AnnChain/vanntool/def"
	"github.com/dappledger/AnnChain/vanntool/tools"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
)

var dm DataManager

type DataManager struct {
	node OperateNode
}

func (m *DataManager) init() {
	m.node.Init()
}

func NodeM() *OperateNode {
	return &dm.node
}

type OperateNode struct {
	o orm.Ormer
}

func (n *OperateNode) Init() {
	n.o = orm.NewOrm()
	n.o.Using("sqlite3")
}

func (n *OperateNode) Privkey(name string) string {
	return n.Get(name).Privkey
}

func (n *OperateNode) DePrivkey(name, pwd string) string {
	pk := n.Privkey(name)
	if len(pk) == 0 {
		return ""
	}
	plainPk, err := tools.DecryptHexText(pk, []byte(pwd))
	if err != nil {
		return ""
	}
	return string(plainPk)
}

func (n *OperateNode) Pubkey(name, pwd string) string {
	plainPk := n.DePrivkey(name, pwd)
	return tools.ED25519Pubkey(string(plainPk))
}

func (n *OperateNode) RPCAddr(name string) string {
	return n.Get(name).RPCAddr
}

func (n *OperateNode) IP(name string) string {
	return n.Get(name).IP
}

func (n *OperateNode) Get(name string) (node NodeData) {
	node.Name = name
	n.o.Read(&node)
	return
}

func (n *OperateNode) GetByRPC(rpc string) (node NodeData) {
	node.RPCAddr = rpc
	n.o.Read(&node, "rpc_addr")
	return
}

func (n *OperateNode) checkParams(name, rpc string) error {
	if !tools.OnlyNumLetterUnderline(name) {
		return fmt.Errorf("node name only support number|letter|_")
	}
	return tools.CheckIPAddr("tcp", rpc)
}

func (n *OperateNode) Insert(name, rpc, pk string) error {
	var err error
	if err = n.checkParams(name, rpc); err != nil {
		return err
	}
	node := NodeData{
		Name:    name,
		RPCAddr: fmt.Sprintf("%v%v", def.TCP_PREFIX, rpc),
		IP:      tools.IPFromAddr(rpc),
		Privkey: pk,
	}
	_, err = n.o.Insert(&node)
	beego.Debug("[operate_node],insert node,name:", name, ",RPCAddr:", rpc, ",Privkey:", pk, ",ip:", node.IP, ",err:", err)
	return err
}

func (n *OperateNode) Drop(name string) error {
	var err error
	node := NodeData{
		Name: name,
	}
	_, err = n.o.Delete(&node)
	beego.Debug("[operate_node],delete node,name:", name, ",err:", err)
	return err
}

func (n *OperateNode) Modify(name, rpc, pk string) error {
	var err error
	if err = n.checkParams(name, rpc); err != nil {
		return err
	}
	node := NodeData{
		Name:    name,
		RPCAddr: fmt.Sprintf("%v%v", def.TCP_PREFIX, rpc),
		IP:      tools.IPFromAddr(rpc),
		Privkey: pk,
	}
	_, err = n.o.Update(&node)
	beego.Debug("[operate_node],modify node,name:", name, ",RPCAddr:", rpc, ",Privkey:", pk, ",err:", err)
	return err
}

func (n *OperateNode) ListNode() []NodeDataShow {
	var nodes []*NodeData
	n.o.QueryTable(&NodeData{}).All(&nodes)
	return NodeSlcToShowSlc(nodes)
}

type NodeInfoShow struct {
	Apps      string `json:"apps"`
	Events    string `json:"events"`
	LastBlock string `json:"last_block"`
	Code      string `json:"code"`
}

func (s *NodeInfoShow) String() string {
	bt, _ := json.Marshal(s)
	return string(bt)
}

func (n *OperateNode) ShowInfo(name, orgname, codeHash string) (show NodeInfoShow, err error) {
	node := n.Get(name)
	if len(node.RPCAddr) == 0 {
		err = fmt.Errorf("not find rpc address of the node")
		return
	}
	var query QueryNodeData
	query.Init(node.RPCAddr, orgname)
	show.Apps = (&query).Apps()
	show.Events = (&query).Events()
	show.LastBlock = (&query).LastBlock()
	show.Code = (&query).EventCode(codeHash)
	return
}
