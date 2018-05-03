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


package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	infoWithDebug struct{}
)

func (l infoWithDebug) Enabled(lv zapcore.Level) bool {
	return lv == zapcore.InfoLevel || lv == zapcore.DebugLevel
}

func init() {
	encoderCfg := zap.NewProductionEncoderConfig()
	zcore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(os.Stdout),
		infoWithDebug{},
	)

	logger = zap.New(zcore)
}
