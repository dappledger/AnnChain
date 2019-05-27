package log

import "testing"

func TestLog(t *testing.T) {
	logger, err := Initialize("dev", "testdir")
	if err != nil {
		t.Error("initialize err ", err)
		return
	}
	SetLog(logger)
	Error("error log")
	Warn("warn log")
	Info("info log")
}
