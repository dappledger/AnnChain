// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	_ "github.com/golang/glog"
	"github.com/dappledger/AnnChain/ann-module/lib/astaxie/beego/logs"
)

// RFC5424 log message levels.
const (
	LevelEmergency     = iota //1
	LevelAlert                //2
	LevelCritical             //3
	LevelError                //4
	LevelWarning              //5
	LevelInformational        //6
	LevelDebug                //7
)

type Logger struct {
	l *logs.BeeLogger
}

var (
	DefaultLogger *Logger
	//DefaultSMTPLogger *Logger
	//DefaultConnLogger *Logger
	//DefaultFileLogger *Logger
	//DefaultGlogLogger *Logger
	DefaultMsgCount int64 = 10000
	DefalutLogLevel       = logs.LevelDebug
)

func newLog(logger, jsonconfig string, bufsize int64) (l *Logger, err error) {
	l = &Logger{l: logs.NewLogger(bufsize)}
	if logger == "" {
		return
	}
	err = l.l.SetLogger(logger, jsonconfig)
	if err != nil {
		return nil, err
	}
	var config struct {
		Level int `json:level`
	}
	err = json.Unmarshal([]byte(jsonconfig), &config)
	if err != nil {
		return nil, err
	}
	l.l.SetLevel(config.Level)

	return l, err
}

func Init(logger, jsonconfig string, async bool) (err error) {
	DefaultLogger, err = newLog(logger, jsonconfig, DefaultMsgCount)
	if err != nil {
		return
	}

	//开启异步
	if async {
		DefaultLogger.l.Async()
	}

	//追踪深度为4
	DefaultLogger.l.SetLogFuncCallDepth(4)
	//开启代码追踪
	DefaultLogger.l.EnableFuncCallDepth(true)

	if logger == "glog" {
		flag.Parse() // necessory for glog
	}
	return
}

func SetLevel(level int) {
	DefaultLogger.SetLevel(level)
}

func Emergency(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Emergency(msg...)
}

func Emergencyf(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Emergencyf(format, msg...)
}

func Alert(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Alert(msg...)
}

func Alertf(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Alertf(format, msg...)
}

func Critical(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Critical(msg...)
}

func Criticalf(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Criticalf(format, msg...)
}

func Error(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Error(msg...)
}

func Errorf(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Errorf(format, msg...)
}

func Warn(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Warn(msg...)
}

func Warnf(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Warnf(format, msg...)
}

//func Notice(msg ...interface{}) {
//	if DefaultLogger == nil {
//		return
//	}
//	DefaultLogger.Notice(msg...)
//}

//func Noticef(format string, msg ...interface{}) {
//	if DefaultLogger == nil {
//		return
//	}
//	DefaultLogger.Noticef(format, msg...)
//}

func Info(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Info(msg...)
}

func Infof(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Infof(format, msg...)
}

func Debug(msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Debug(msg...)
}

func Debugf(format string, msg ...interface{}) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Debugf(format, msg...)
}

//func Detail(msg ...interface{}) {
//	if DefaultLogger == nil {
//		return
//	}
//	DefaultLogger.Detail(msg...)
//}

//func Detailf(format string, msg ...interface{}) {
//	if DefaultLogger == nil {
//		return
//	}
//	DefaultLogger.Detailf(format, msg...)
//}

func Flush() {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Flush()
}

func Close() {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Close()
}

func Exit(code int) {
	runtime.Gosched()
	os.Exit(code)
}

func (l *Logger) SetLevel(level int) {
	l.l.SetLevel(level)
}

func (l *Logger) Emergency(msg ...interface{}) {
	l.l.Emergency("%+v\n", msg)
}

func (l *Logger) Emergencyf(format string, msg ...interface{}) {
	l.l.Emergency(format, msg...)
}

func (l *Logger) Alert(msg ...interface{}) {
	l.l.Alert("%+v", msg)
}

func (l *Logger) Alertf(format string, msg ...interface{}) {
	l.l.Alert(format, msg...)
}

func (l *Logger) Critical(msg ...interface{}) {
	l.l.Critical("%+v", msg)
}

func (l *Logger) Criticalf(format string, msg ...interface{}) {
	l.l.Critical(format, msg...)
}

func (l *Logger) Error(msg ...interface{}) {
	l.l.Error("%+v", msg)
}

func (l *Logger) Errorf(format string, msg ...interface{}) {
	l.l.Error(format, msg...)
}

func (l *Logger) Warn(msg ...interface{}) {
	l.l.Warn("%+v", msg)
}

func (l *Logger) Warnf(format string, msg ...interface{}) {
	l.l.Warn(format, msg...)
}

//func (l *Logger) Notice(msg ...interface{}) {
//	l.l.Notice("%+v", msg)
//}

//func (l *Logger) Noticef(format string, msg ...interface{}) {
//	l.l.Notice(format, msg...)
//}

func (l *Logger) Info(msg ...interface{}) {
	l.l.Info("%+v", msg)
}

func (l *Logger) Infof(format string, msg ...interface{}) {
	l.l.Info(format, msg...)
}

func (l *Logger) Debug(msg ...interface{}) {
	fmt.Println("----------------------Debug------------------------")
	l.l.Debug("%+v", msg)
}

func (l *Logger) Debugf(format string, msg ...interface{}) {
	l.l.Debug(format, msg...)
}

//func (l *Logger) Detail(msg ...interface{}) {
//	l.l.Detail("%+v\n", msg)
//}

//func (l *Logger) Detailf(format string, msg ...interface{}) {
//	l.l.Detail(format, msg...)
//}

func (l *Logger) Flush() {
	l.l.Flush()
}

// Close the logger and waiting all messages was printed
func (l *Logger) Close() {
	l.Flush()
	l.l.Close()
}
