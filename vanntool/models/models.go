package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"

	"github.com/dappledger/AnnChain/vanntool/def"
	"github.com/dappledger/AnnChain/vanntool/tools"
)

type CmdItfc interface {
	ServeCmdItfc
	FillData(*beego.Controller) error
	Do() string
}

type ServeCmdItfc interface {
	Args() []string
}

func GetCmdOp(cmd, op string) CmdItfc {
	var do CmdItfc = nil
	switch cmd {
	case "sign":
		do = &SignFull{}
	case "special":
		switch op {
		case "change_validator":
			do = &SpecialChangeValidatorFull{}
		}
	case "organization":
		switch op {
		case "create":
			do = &OrgCreateFull{}
		case "join":
			do = &OrgJoinFull{}
		case "leave":
			do = &OrgLeaveFull{}
		}
	case "event":
		switch op {
		case "uploadcode":
			do = &EventUploadCodeFull{}
		case "request":
			do = &EventRequestFull{}
		case "unsubscribe":
			do = &EventUnsubscribeFull{}
		}
	case "evm":
		switch op {
		case "create":
			do = &CreateContractFull{}
		case "call":
			fallthrough
		case "read":
			do = &CallOrReadContractFull{
				CallOrReadContract: CallOrReadContract{
					Op: op,
				},
			}
		}
	case "jvm":
		switch op {
		case "create":
			do = &CreateJvmContractFull{}
		case "call", "read":
			do = &CallOrQueryJvmContractFull{
				CallOrQueryJvmContract: CallOrQueryJvmContract{
					Op: op,
				},
			}
		}
	}
	return do
}

func ServeCmd(cmd ServeCmdItfc) string {
	return RunShell(cmd.Args())
}

var AlreadyInTesting bool

func BatchServCmd(cmd ServeCmdItfc, roNum, numPerSec uint) error {
	if AlreadyInTesting {
		return nil
	}
	args := cmd.Args()
	RunShell(args)
	return nil
}

type Base struct {
	BackEnd string `form:"backend"`
	Target  string `form:"target"`
}

func (b *Base) BaseArgs() []string {
	return ParseArgs(b, nil)
}

func parseIPAddr(backend string) (bk string) {
	if !strings.Contains(backend, ":") {
		node := NodeM().Get(backend)
		return node.RPCAddr
	}
	if strings.HasPrefix(backend, def.TCP_PREFIX) {
		backend = backend[len(def.TCP_PREFIX):]
	}
	if err := tools.CheckIPAddr("tcp", backend); err != nil {
		fmt.Println("err", err)
		return
	}
	bk = fmt.Sprintf("%v%v", def.TCP_PREFIX, backend)
	return
}

// parseNodePrivkey returns the real rpc address and privkey of the node.
// If 'backend' is the name of the node,
// then 'privkey' is the pwd of AES-encrypto of the node's privkey.
// Or just literally meaning.
func parseNodePrivkey(backend, privkey string) (bk, pk string) {
	if strings.Index(backend, ":") > -1 {
		// this way,backend can't be node_name
		bk = parseIPAddr(backend)
		if len(bk) == 0 {
			return
		}
		pk = privkey
		return
	}
	node := NodeM().Get(backend)
	if len(node.Privkey) != 0 {
		if debytes, err := tools.DecryptHexText(node.Privkey, []byte(privkey)); err == nil {
			pk = string(debytes)
			bk = node.RPCAddr
		}
	}
	return
}

func parseNodeNamePwd(backend, privkey string) (nodeName, nodePwd string) {
	if index := strings.Index(backend, ":"); index == -1 {
		nodeName = backend
		nodePwd = privkey
	}
	return
}
