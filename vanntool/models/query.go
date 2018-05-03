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

type QueryNodeData struct {
	Base
}

func (q *QueryNodeData) Init(backend, orgid string) {
	q.BackEnd = backend
	q.Target = orgid
}

func (q *QueryNodeData) Apps() string {
	return RunShell(ParseArgs(q, append(q.BaseArgs(), "query", "apps")))
}

func (q *QueryNodeData) Events() string {
	if len(q.Target) == 0 {
		return ""
	}
	return RunShell(ParseArgs(q, append(q.BaseArgs(), "query", "events")))
}

func (q *QueryNodeData) LastBlock() string {
	if len(q.Target) == 0 {
		return ""
	}
	return RunShell(ParseArgs(q, append(q.BaseArgs(), "info", "last_block")))
}

func (q *QueryNodeData) EventCode(codeHash string) string {
	if len(q.Target) == 0 || len(codeHash) == 0 {
		return ""
	}
	var ec QueryEventCode
	ec.QueryNodeData = *q
	ec.CodeHash = codeHash
	return RunShell(ParseArgs(&ec, append(ec.BaseArgs(), "query", "event_code")))
}

type QueryEventCode struct {
	QueryNodeData
	CodeHash string `form:"code_hash"`
}
