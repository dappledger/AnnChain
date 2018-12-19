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
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZap(t *testing.T) {

	env := "production"

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

	logger := zap.New(zapcore.NewTee(coreInfo, coreError))

	logger.Info("11111111")
	logger.Info("2222222")
	logger.Error("EEEEEEEEEE")
	logger.Info("3333333")
	logger.Info("44444444")
}
