package models

import (
	"strings"

	"github.com/astaxie/beego"
)

type SignFull struct {
	Sign
	BackEnd string `form:"backend"`
}

func (f *SignFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	f.BackEnd, f.Sign.Sec = parseNodePrivkey(f.BackEnd, f.Sign.Sec)
	return nil
}

type Sign struct {
	Sec string `form:"sec"`
	Pub string `form:"pub"`
}

func (s *Sign) Args() []string {
	//vl := reflect.ValueOf(s).Elem()
	//args := ParseArgs(reflect.TypeOf(s).Elem(), &vl)
	//return append([]string{"sign"}, args...)
	return ParseArgs(s, []string{"sign"})
}

func (s *Sign) Do() string {
	return ServeCmd(s)
}

func (s *Sign) DoSign(str string) string {
	cps := *s
	cps.Pub = str
	ret := RunShell(cps.Args())
	if idx := strings.Index(ret, ":"); idx > 0 && idx < len(ret)-1 {
		return strings.TrimSpace(ret[idx+1:])
	}
	return ""
}
