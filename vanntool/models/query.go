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
