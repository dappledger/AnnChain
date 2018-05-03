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
	"path"

	"github.com/utahta/go-cronowriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitializeLog(env, logpath string) *zap.Logger {
	if env == "production" {
		jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		cw := cronowriter.MustNew(path.Join(logpath, "output.log.%Y%m%d"))

		prdCore := zapcore.NewCore(
			jsonEncoder,
			zapcore.AddSync(cw),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.InfoLevel
			}))

		return zap.New(prdCore)
	}

	zapConf := zap.NewDevelopmentConfig()
	zapConf.OutputPaths = []string{path.Join(logpath, "output.log")}

	// zapConf.ErrorOutputPaths = []string{path.Join(logpath, "err.output.log")}
	if logger, err := zapConf.Build(); err == nil {
		return logger
	} else {
		panic(err)
	}

}
