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

	"go.uber.org/zap"
)

func InitializeLog(env, logpath string) *zap.Logger {
	var zapConf zap.Config
	var err error

	if env == "production" {
		zapConf = zap.NewProductionConfig()
	} else {
		zapConf = zap.NewDevelopmentConfig()
	}

	zapConf.OutputPaths = []string{path.Join(logpath, "output.log")}
	zapConf.ErrorOutputPaths = []string{path.Join(logpath, "err.output.log")}
	logger, err := zapConf.Build()
	if err != nil {
		panic(err.Error())
	}

	return logger
}
