package commands

import (
	"os"
	"path"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	zapConf := zap.NewDevelopmentConfig()
	zapConf.OutputPaths = []string{path.Join(pwd, "client.out.log")}
	zapConf.ErrorOutputPaths = []string{}
	logger, err = zapConf.Build()
	if err != nil {
		panic(err.Error())
	}
}
