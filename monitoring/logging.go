package monitoring

import "go.uber.org/zap"

var logger *zap.SugaredLogger

func InitLogging() {
	zl, _ := zap.NewProduction()
	logger = zl.Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
