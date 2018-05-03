package models

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	agtypes "github.com/dappledger/AnnChain/angine/types"
	"github.com/dappledger/AnnChain/module/lib/go-crypto"
	"github.com/dappledger/AnnChain/vanntool/tools"
)

const (
	SPLIT_SYMBOL = "?"
)

func orgToSign(nodeName, nodePwd, orgName string) (toSign string) {
	if len(nodeName) == 0 || len(nodePwd) == 0 {
		return
	}
	pubk := NodeM().Pubkey(nodeName, nodePwd)
	if len(pubk) == 0 {
		return
	}
	return pubk + orgName
}

func namePortToAddr(nodeName, port string) (addr string) {
	ip := NodeM().IP(nodeName)
	if len(ip) == 0 {
		return
	}
	if len(port) == 0 {
		return
	}
	return fmt.Sprintf("%v:%v", ip, port)
}

type DoWithInputFunc func(string) string

// oriSeeds:<IP>:<Port> || <节点名>:<Port> 统一输出 <IP>:<Port>
func parseSeeds(oriSeeds string) (seeds string) {
	if len(oriSeeds) == 0 {
		return
	}
	ipSlc := strings.Split(oriSeeds, ",")
	var newSeeds string
	for i := range ipSlc {
		addr := ipSlc[i]
		if err := tools.CheckIPAddr("tcp", addr); err != nil {
			index := strings.Index(oriSeeds, ":")
			if index == -1 || index >= len(oriSeeds)-1 {
				return
			}
			addr = namePortToAddr(ipSlc[i][:index], ipSlc[i][index+1:])
			if len(addr) == 0 {
				return
			}
		}
		if len(newSeeds) == 0 {
			newSeeds = addr
		} else {
			newSeeds = fmt.Sprintf("%v,%v", newSeeds, addr)
		}
	}
	seeds = newSeeds
	return
}

// p2p_laddr: tcp://<IP>:<Port> || <Port> 统一输出 tcp://<IP>:<Port>
func parseP2PLaddr(p2p_laddr string, doWithPort DoWithInputFunc) (addr string) {
	if len(p2p_laddr) == 0 {
		return
	}
	newAddr := p2p_laddr
	if err := tools.CheckIPAddr("tcp", newAddr); err != nil {
		newAddr = doWithPort(p2p_laddr)
	}
	return newAddr
}

func splitNamePwd(input string) (name, pwd string) {
	index := strings.Index(input, SPLIT_SYMBOL)
	if index > -1 {
		if index >= len(input)-1 {
			return
		}
		name = input[:index]
		pwd = input[index+1:]
	}
	return
}

// signByCA: <signedSig> || <CANode_Name>?<CANode_pwd>
func parseCANode(signByCA string, doWithPriv DoWithInputFunc) (caSign string) {
	if len(signByCA) == 0 {
		return
	}
	name, pwd := splitNamePwd(signByCA)
	if len(name) > 0 {
		if doWithPriv == nil {
			return
		}
		rpk := NodeM().DePrivkey(name, pwd)
		caSign = doWithPriv(rpk)
	} else {
		caSign = signByCA
	}
	return
}

var (
	E_PARSE_GENESFILE      = errors.New("genesfile parse error")
	E_NODE_NOT_FOUND       = errors.New("node not found")
	E_VALIDATOR_PARAM_ANAL = errors.New("analysis validator params error")
)

// parse string: amount:100,is_ca:true
func parseValidatorParams(params string, v *agtypes.GenesisValidator) error {
	pslc := strings.Split(params, ",")
	if len(pslc) > 4 {
		return E_VALIDATOR_PARAM_ANAL
	}
	for i := range pslc {
		name, param := tools.SplitTo2(pslc[i], ":")
		name = strings.Trim(name, " ")
		param = strings.Trim(param, " ")
		switch name {
		case "amount":
			a, err := strconv.Atoi(param)
			if err != nil {
				return fmt.Errorf("%v:%v", E_VALIDATOR_PARAM_ANAL, err)
			}
			v.Amount = int64(a)
		case "is_ca":
			b, err := strconv.ParseBool(param)
			if err != nil {
				return fmt.Errorf("%v:%v", E_VALIDATOR_PARAM_ANAL, err)
			}
			v.IsCA = b
		default:
			return fmt.Errorf("%v:%v", E_VALIDATOR_PARAM_ANAL, name)
		}
	}
	return nil
}

func parseGenesisNodeToFile(orgid, gennode string) (string, error) {
	if len(gennode) == 0 {
		return "", nil
	}
	nodes := strings.Split(gennode, ";")
	var genisNode agtypes.GenesisDoc
	genisNode.ChainID = orgid
	genisNode.GenesisTime = agtypes.Time{time.Now()}
	genisNode.Plugins = "specialop"
	genisNode.Validators = make([]agtypes.GenesisValidator, 0, len(nodes))
	for i := range nodes {
		if len(nodes[i]) == 0 {
			return "", fmt.Errorf("node[%v],err:%v", i, E_PARSE_GENESFILE)
		}
		index := strings.Index(nodes[i], "(")
		nodeInfo := nodes[i]
		var v agtypes.GenesisValidator
		if index >= 0 && index != len(nodes[i])-1 {
			indexEnd := strings.Index(nodes[i][index:], ")")
			if indexEnd < 0 {
				return "", fmt.Errorf("node[%v],')' err:%v", i, E_PARSE_GENESFILE)
			}
			vstr := nodes[i][index+1 : index+indexEnd]
			if err := parseValidatorParams(vstr, &v); err != nil {
				return "", err
			}
			nodeInfo = nodes[i][:index]
		} else {
			index = len(nodes[i]) - 1
		}
		name, pwd := splitNamePwd(nodeInfo)
		if len(name) == 0 {
			return "", fmt.Errorf("node[%v],name pwd err:%v", i, E_PARSE_GENESFILE)
		}
		pubkey := NodeM().Pubkey(name, pwd)
		if len(pubkey) == 0 {
			return "", errors.New(fmt.Sprintf("%v:%v", name, E_NODE_NOT_FOUND))
		}
		pubkeyBytes, _ := hex.DecodeString(pubkey)
		var pArray [32]byte
		copy(pArray[:], pubkeyBytes)
		pubkey25519 := crypto.PubKeyEd25519(pArray)
		v.PubKey = crypto.StPubKey{&pubkey25519}
		genisNode.Validators = append(genisNode.Validators, v)
	}
	//ret, err := json.Marshal(&genisNode)
	ret, err := genisNode.JSONBytes()
	if err != nil {
		beego.Warn("[parse_genes],gen genesis file err:", err)
		return "", err
	}
	return string(ret), nil
}
