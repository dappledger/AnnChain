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
