package log

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

func Initialize(mode, env, output, errOutput string) *zap.Logger {
	var logI *zap.Logger

	if mode == "file" {
		var zapConf zap.Config
		var err error

		if env == "production" {
			zapConf = zap.NewProductionConfig()
		} else {
			zapConf = zap.NewDevelopmentConfig()
		}

		zapConf.OutputPaths = []string{output}
		zapConf.ErrorOutputPaths = []string{errOutput}
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
