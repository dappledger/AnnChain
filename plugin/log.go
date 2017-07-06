package plugin

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func SetLog(l *zap.Logger) {
	log = l
}
