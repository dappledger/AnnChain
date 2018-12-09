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


package log

import (
	"path"

	"go.uber.org/zap"

	cmn "github.com/dappledger/AnnChain/module/lib/go-common"
)

func Initialize(env, output, errOutput string) *zap.Logger {
	var zapConf zap.Config
	var err error

	if env == "production" {
		zapConf = zap.NewProductionConfig()
	} else {
		zapConf = zap.NewDevelopmentConfig()
	}

	cmn.EnsureDir(path.Dir(output), 0775)
	cmn.EnsureDir(path.Dir(errOutput), 0775)

	zapConf.OutputPaths = []string{output}
	zapConf.ErrorOutputPaths = []string{errOutput}
	logger, err := zapConf.Build()
	if err != nil {
		panic(err.Error()) // which should never happen
	}

	logger.Debug("Starting zap! Have your fun!")

	return logger
}
