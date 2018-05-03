package models

import (
	"fmt"

	"github.com/dappledger/AnnChain/vanntool/def"
	"github.com/dappledger/AnnChain/vanntool/tools"

	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego"
)

type OrgExtra struct {
	AppName   string `form:"app_list"`
	P2P_laddr string `form:"p2p_laddr"`
	Seeds     string `form:"seeds"`
	CANode    string `form:"sign_by_CA"`
}

func (oe *OrgExtra) parse(nodeName, nodePwd, orgId string) (changed bool, err error) {
	if len(nodeName) == 0 {
		return
	}
	var seeds, p2p_laddr, sign_by_CA string
	if len(oe.Seeds) > 0 {
		seeds = parseSeeds(oe.Seeds)
		oe.Seeds = seeds
		changed = true
	}
	if len(oe.P2P_laddr) > 0 {
		p2p_laddr = parseP2PLaddr(oe.P2P_laddr, func(port string) (addr string) {
			return fmt.Sprintf("%v0.0.0.0:%v", def.TCP_PREFIX, port)
		})
		if len(p2p_laddr) == 0 {
			err = fmt.Errorf("can't parse p2p_laddr")
			return
		}
		oe.P2P_laddr = p2p_laddr
		changed = true
	}
	if len(oe.CANode) > 0 {
		sign_by_CA = parseCANode(oe.CANode, func(priv string) (caSign string) {
			plainText := orgToSign(nodeName, nodePwd, orgId)
			if len(plainText) == 0 {
				return
			}
			var s Sign
			s.Sec = priv
			return s.DoSign(plainText)
		})
		if len(sign_by_CA) == 0 {
			err = fmt.Errorf("can't parse p2p_laddr,can't find the node or parse err")
			return
		}
		oe.CANode = sign_by_CA
		changed = true
	}
	return
}

func (oe *OrgExtra) filterConfMap(confMap map[string]interface{}) {
	if len(oe.Seeds) != 0 {
		confMap["seeds"] = oe.Seeds
	}
	if len(oe.P2P_laddr) != 0 {
		confMap["p2p_laddr"] = oe.P2P_laddr
	}
	if len(oe.CANode) != 0 {
		confMap["signbyCA"] = oe.CANode
	}
	if len(oe.AppName) != 0 {
		confMap["appname"] = oe.AppName
	}
}

//////////////////////////////////////////////////////////////////////////////////

type OrgCreate struct {
	Base
	Privkey     string `form:"privkey"`
	GenesisFile string `form:"genesisfile"`
	ConfigFile  string `form:"configfile"`
	parsed      bool
}

func (scv *OrgCreate) parseFile() {
	if !scv.parsed {
		scv.GenesisFile = ParseStringArg(scv.GenesisFile)
		scv.ConfigFile = ParseStringArg(scv.ConfigFile)
		scv.parsed = true
	}
}

func (scv *OrgCreate) Args() []string {
	return ParseArgs(scv, append(scv.BaseArgs(), []string{"organization", "create"}...))
}

type OrgCreateFull struct {
	OrgCreate
	OrgExtra
	Orgid       string `form:"orgid"`
	GenesisNode string `form:"genesisnode"`
}

func (f *OrgCreateFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	nodeName, nodePwd := parseNodeNamePwd(f.OrgCreate.BackEnd, f.OrgCreate.Privkey)
	f.OrgCreate.BackEnd, f.OrgCreate.Privkey = parseNodePrivkey(f.OrgCreate.BackEnd, f.OrgCreate.Privkey)
	if len(f.OrgCreate.BackEnd) == 0 || len(f.OrgCreate.Privkey) == 0 {
		return fmt.Errorf("backend || privkey == nil,err")
	}

	confMap := make(map[string]interface{})
	var err error
	// err = toml.Unmarshal([]byte(f.ConfigFile), &confMap)
	err = toml.Unmarshal([]byte(f.ConfigFile), &confMap)
	if err != nil {
		beego.Warn("[org_create],Unmarshal ConfigFile err:", err)
		return fmt.Errorf("unmarshal configfile err,%v", err.Error())
	}
	var changed bool
	changed, err = f.parse(nodeName, nodePwd, f.Orgid)
	if err != nil {
		return err
	}
	if changed {
		f.filterConfMap(confMap)
	}
	if len(f.GenesisNode) > 0 {
		f.GenesisFile, err = parseGenesisNodeToFile(f.Orgid, f.GenesisNode)
		if err != nil {
			return err
		}
	}
	err = checkConfMap(confMap)
	if err != nil {
		return err
	}
	var confRet string
	confRet, err = tools.EncodeToToml(&confMap)
	if err != nil {
		return err
	}
	f.ConfigFile = confRet
	return nil
}

func checkConfMap(cmp map[string]interface{}) error {
	if app, ok := cmp["appname"]; ok {
		switch app {
		case "remote":
			if _, ok := cmp["rpcapp_addr"]; !ok {
				return fmt.Errorf("remote app need param of rpcapp_addr(<IP>:<Port>)")
			}
		case "evm":
			if _, ok := cmp["cosi_laddr"]; !ok {
				return fmt.Errorf("evm app need param of cosi_laddr(tcp://<IP>:<Port>)")
			}
			if _, ok := cmp["event_laddr"]; !ok {
				return fmt.Errorf("evm app need param of event_laddr(tcp://<IP>:<Port>)")
			}
		case "ikhofi":
			if _, ok := cmp["ikhofi_addr"]; !ok {
				return fmt.Errorf("ikhofi app need param of ikhofi_addr(http://<IP>:<Port>)")
			}
			if _, ok := cmp["cosi_laddr"]; !ok {
				return fmt.Errorf("ikhofi app need param of cosi_laddr(tcp://<IP>:<Port>)")
			}
		}
	}
	return nil
}

func (f *OrgCreateFull) Do() string {
	return ServeCmd(f)
}

//////////////////////////////////////////////////////////////////////////////////

type OrgJoin struct {
	Base
	Orgid      string `form:"orgid"`
	Privkey    string `form:"privkey"`
	ConfigFile string `form:"configfile"`
	//GenesisFile string `form:"genesisfile"`
	parsed bool
}

func (scv *OrgJoin) parseFile() {
	if !scv.parsed {
		scv.ConfigFile = ParseStringArg(scv.ConfigFile)
		scv.parsed = true
	}
}

func (scv *OrgJoin) Args() []string {
	return ParseArgs(scv, append(scv.BaseArgs(), []string{"organization", "join"}...))
}

type OrgJoinFull struct {
	OrgJoin
	OrgExtra
}

func (f *OrgJoinFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	nodeName, nodePwd := parseNodeNamePwd(f.OrgJoin.BackEnd, f.OrgJoin.Privkey)
	f.OrgJoin.BackEnd, f.OrgJoin.Privkey = parseNodePrivkey(f.OrgJoin.BackEnd, f.OrgJoin.Privkey)
	if len(f.OrgJoin.BackEnd) == 0 || len(f.OrgJoin.Privkey) == 0 {
		return fmt.Errorf("backend || privkey == nil,err")
	}

	confMap := make(map[string]interface{})
	var err error
	err = toml.Unmarshal([]byte(f.ConfigFile), &confMap)
	if err != nil {
		beego.Warn("[org_create],Unmarshal ConfigFile err:", err)
		return fmt.Errorf("unmarshal configfile err,%v", err.Error())
	}
	var changed bool
	changed, err = f.parse(nodeName, nodePwd, f.Orgid)
	if err != nil {
		return err
	}
	if changed {
		f.filterConfMap(confMap)
	}
	err = checkConfMap(confMap)
	if err != nil {
		return err
	}
	var confRet string
	confRet, err = tools.EncodeToToml(&confMap)
	if err != nil {
		return err
	}
	f.ConfigFile = confRet
	return nil
}

func (f *OrgJoinFull) Do() string {
	return ServeCmd(f)
}

//////////////////////////////////////////////////////////////////////////////////

type OrgLeave struct {
	Base
	Privkey string `form:"privkey"`
	Orgid   string `form:"orgid"`
}

func (sl *OrgLeave) Args() []string {
	return ParseArgs(sl, append(sl.BaseArgs(), []string{"organization", "leave"}...))
}

type OrgLeaveFull struct {
	OrgLeave
}

func (f *OrgLeaveFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	f.OrgLeave.BackEnd, f.OrgLeave.Privkey = parseNodePrivkey(f.OrgLeave.BackEnd, f.OrgLeave.Privkey)
	if len(f.OrgLeave.BackEnd) == 0 || len(f.OrgLeave.Privkey) == 0 {
		return fmt.Errorf("backend || privkey == nil,err")
	}
	if len(f.OrgLeave.Orgid) == 0 {
		return fmt.Errorf("lack of orgid")
	}
	return nil
}

func (f *OrgLeaveFull) Do() string {
	return ServeCmd(f)
}
