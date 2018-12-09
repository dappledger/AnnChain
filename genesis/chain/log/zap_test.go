package log

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// type

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
