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

package angine

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	infoOnly      struct{}
	infoWithDebug struct{}
	aboveWarn     struct{}
)

func (l infoOnly) Enabled(lv zapcore.Level) bool {
	return lv == zapcore.InfoLevel
}
func (l infoWithDebug) Enabled(lv zapcore.Level) bool {
	return lv == zapcore.InfoLevel || lv == zapcore.DebugLevel
}
func (l aboveWarn) Enabled(lv zapcore.Level) bool {
	return lv >= zapcore.WarnLevel
}

func makeInfoFilter(env string) zapcore.LevelEnabler {
	switch env {
	case "production":
		return infoOnly{}
	default:
		return infoWithDebug{}
	}
}

func makeErrorFilter() zapcore.LevelEnabler {
	return aboveWarn{}
}

func InitializeLog(mode, env, logpath string) *zap.Logger {
	var logI *zap.Logger

	if mode == "file" {
		var zapConf zap.Config
		var err error

		if env == "production" {
			zapConf = zap.NewProductionConfig()
		} else {
			zapConf = zap.NewDevelopmentConfig()
		}

		zapConf.OutputPaths = []string{"output.log"}
		zapConf.ErrorOutputPaths = []string{"err.output.log"}
		logI, err = zapConf.Build()
		if err != nil {
			panic(err.Error())
		}
	} else {
		var encoderCfg zapcore.EncoderConfig
		if env == "production" {
			encoderCfg = zap.NewProductionEncoderConfig()
		} else {
			encoderCfg = zap.NewDevelopmentEncoderConfig()
		}

		coreInfo := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.NewMultiWriteSyncer(os.Stdout),
			makeInfoFilter(env),
		)
		coreError := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.NewMultiWriteSyncer(os.Stderr),
			makeErrorFilter(),
		)

		logI = zap.New(zapcore.NewTee(coreInfo, coreError))
	}

	return logI
}
