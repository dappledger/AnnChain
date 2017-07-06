package consensus

import (
	"go.uber.org/zap"
)

var log *zap.Logger
var slog *zap.SugaredLogger

func SetLog(l *zap.Logger) {
	log = l
	slog = l.Sugar()
}
