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

package log

import (
	"fmt"
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger
var slogger *zap.SugaredLogger
var auditLogger *zap.Logger

func ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return fmt.Errorf("Could not create directory%v. %v", dir, err)
		}
	}
	return nil
}

func productionLogger(logPath string) *zap.Logger {
	zapEncodeConfig := productionConfig()
	jsonEncoder := zapcore.NewJSONEncoder(zapEncodeConfig)

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    512, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewCore(
		jsonEncoder,
		w,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel
		}),
	)
	return zap.New(core)
}

func productionConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     productionTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func productionTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
}

func Initialize(env, logpath string) (*zap.Logger, error) {
	if err := ensureDir(path.Dir(logpath)); err != nil {
		return nil, err
	}

	if logpath == "" {
		logpath = "output.log"
	}
	if env == "production" {
		return productionLogger(logpath), nil
	}
	zapConf := zap.NewDevelopmentConfig()
	zapConf.OutputPaths = []string{logpath}
	opt := zap.AddCallerSkip(1)
	return zapConf.Build(opt)
}

func SetLog(l *zap.Logger) {
	logger = l
	slogger = l.Sugar()
}

func SetAuditLog(l *zap.Logger) {
	auditLogger = l
}

func setDefaultAuditLog() {
	if auditLogger != nil {
		return
	}
	zapEncodeConfig := zap.NewProductionEncoderConfig()
	jsonEncoder := zapcore.NewJSONEncoder(zapEncodeConfig)

	w := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(
		jsonEncoder,
		w,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel
		}),
	)
	auditLogger = zap.New(core)
}

func InitAuditLog(logPath string) (*zap.Logger, error) {
	if err := ensureDir(path.Dir(logPath)); err != nil {
		return nil, err
	}
	if logPath == "" {
		logPath = "audit.log"
	}
	return productionLogger(logPath), nil
}

func Audit() (l *zap.Logger) {
	//for tests that didn't initialize audit logs
	if auditLogger == nil {
		setDefaultAuditLog()
	}
	return auditLogger
}

func Info(msg string, fields ...zapcore.Field) {
	if logger == nil {
		fmt.Println("[1],log leaks:", msg, fields)
		return
	}
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	if logger == nil {
		fmt.Println("[2],log leaks:", msg, fields)
		return
	}
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	if logger == nil {
		fmt.Println("[3],log leaks:", msg, fields)
		return
	}
	logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	if logger == nil {
		fmt.Println("[4],log leaks:", msg, fields)
		return
	}
	logger.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	if logger == nil {
		fmt.Println("[4.1],log leaks:", msg, fields)
		return
	}
	logger.Fatal(msg, fields...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	if slogger == nil {
		fmt.Println("[5],log leaks:", msg, keysAndValues)
		return
	}
	slogger.Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	if slogger == nil {
		fmt.Println("[6],log leaks:", msg, keysAndValues)
		return
	}
	slogger.Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	if slogger == nil {
		fmt.Println("[7],log leaks:", msg, keysAndValues)
		return
	}
	slogger.Errorw(msg, keysAndValues...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	if slogger == nil {
		fmt.Println("[8],log leaks:", msg, keysAndValues)
		return
	}
	slogger.Debugw(msg, keysAndValues...)
}

func Infof(template string, args ...interface{}) {
	if slogger == nil {
		fmt.Println("[9],log leaks:", template, args)
		return
	}
	slogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	if slogger == nil {
		fmt.Println("[10],log leaks:", template, args)
		return
	}
	slogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	if slogger == nil {
		fmt.Println("[11],log leaks:", template, args)
		return
	}
	slogger.Errorf(template, args...)
}

func Debugf(template string, args ...interface{}) {
	if slogger == nil {
		fmt.Println("[12],log leaks:", template, args)
		return
	}
	slogger.Debugf(template, args...)
}
