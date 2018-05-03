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
	"fmt"

	"github.com/astaxie/beego"
	cvtools "github.com/dappledger/AnnChain/src/tools"
)

//////////////////////////////////////////////////////////////////////////////////

type EventUploadCode struct {
	Base
	Privkey string `form:"privkey"`
	Code    string `form:"code"`

	// TODO 目前貌似code没有owner，代码只是部署在基链，调用通过hash(code)
	Ownerid string `form:"ownerid"`

	parsed bool
}

func (scv *EventUploadCode) Args() []string {
	return ParseArgs(scv, append(scv.BaseArgs(), []string{"event", "upload-code"}...))
}

type EventUploadCodeFull struct {
	EventUploadCode
	CodeInput string `form:"code_text"`
}

func (f *EventUploadCodeFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	f.EventUploadCode.BackEnd, f.EventUploadCode.Privkey = parseNodePrivkey(f.EventUploadCode.BackEnd, f.EventUploadCode.Privkey)
	if len(f.EventUploadCode.BackEnd) == 0 || len(f.EventUploadCode.Privkey) == 0 {
		return fmt.Errorf("backend || privkey == nil,err")
	}
	if err := cvtools.LuaSyntaxCheck(f.CodeInput); err != nil {
		return err
	}
	f.Code = f.CodeInput
	return nil
}

func (f *EventUploadCodeFull) Do() string {
	return ServeCmd(f)
}

//////////////////////////////////////////////////////////////////////////////////

type EventRequest struct {
	Base
	Privkey      string `form:"privkey"`
	Listener     string `form:"listener"`
	ListenerHash string `form:"listener_hash"`
	Source       string `form:"source"`
	SourceHash   string `form:"source_hash"`

	parsed bool
}

func (scv *EventRequest) Args() []string {
	return ParseArgs(scv, append(scv.BaseArgs(), []string{"event", "request"}...))
}

type EventRequestFull struct {
	EventRequest
}

func (f *EventRequestFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	f.EventRequest.BackEnd, f.EventRequest.Privkey = parseNodePrivkey(f.EventRequest.BackEnd, f.EventRequest.Privkey)
	if len(f.EventRequest.BackEnd) == 0 || len(f.EventRequest.Privkey) == 0 {
		return fmt.Errorf("backend || privkey == nil,err")
	}
	return nil
}

func (f *EventRequestFull) Do() string {
	return ServeCmd(f)
}

//////////////////////////////////////////////////////////////////////////////////

type EventUnsubscribe struct {
	Base
	Privkey  string `form:"privkey"`
	Listener string `form:"listener"`
	Event    string `form:"source"`

	parsed bool
}

func (scv *EventUnsubscribe) Args() []string {
	return ParseArgs(scv, append(scv.BaseArgs(), []string{"event", "request"}...))
}

type EventUnsubscribeFull struct {
	EventUnsubscribe
}

func (f *EventUnsubscribeFull) FillData(c *beego.Controller) error {
	c.ParseForm(f)
	f.EventUnsubscribe.BackEnd, f.EventUnsubscribe.Privkey = parseNodePrivkey(f.EventUnsubscribe.BackEnd, f.EventUnsubscribe.Privkey)
	if len(f.EventUnsubscribe.BackEnd) == 0 || len(f.EventUnsubscribe.Privkey) == 0 {
		return fmt.Errorf("backend || privkey == nil,err")
	}
	return nil
}

func (f *EventUnsubscribeFull) Do() string {
	return ServeCmd(f)
}
