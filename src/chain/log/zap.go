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
