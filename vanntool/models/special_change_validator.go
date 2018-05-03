package models

import (
	"fmt"

	"github.com/astaxie/beego"
)

type SpecialChangeValidator struct {
	Base
	Privkey  string `form:"privkey"`
	NVPubkey string `form:"validator_pubkey"`
	Power    uint   `form:"power"`
	IsCA     bool   `form:"isCA"`
}

func (srv *SpecialChangeValidator) Args() []string {
	return ParseArgs(srv, append(srv.BaseArgs(), []string{"special", "change_validator"}...))
}

func (srv *SpecialChangeValidator) Do() string {
	return ServeCmd(srv)
}

type SpecialChangeValidatorFull struct {
	SpecialChangeValidator
	ToVNode string `form:"to_v_node"`
}

func (f *SpecialChangeValidatorFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	f.SpecialChangeValidator.BackEnd, f.SpecialChangeValidator.Privkey = parseNodePrivkey(f.SpecialChangeValidator.BackEnd, f.SpecialChangeValidator.Privkey)
	if len(f.SpecialChangeValidator.Privkey) == 0 {
		return fmt.Errorf("not find privkey")
	}
	nvName, nvPwd := splitNamePwd(f.ToVNode)
	f.NVPubkey = f.ToVNode
	if len(nvName) > 0 {
		f.NVPubkey = NodeM().Pubkey(nvName, nvPwd)
	}
	if len(f.NVPubkey) == 0 {
		return fmt.Errorf("can't find pubkey of %v", nvName)
	}
	return nil
}
