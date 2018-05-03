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
