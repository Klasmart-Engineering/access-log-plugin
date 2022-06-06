package logging

import (
	lura "github.com/luraproject/lura/v2/logging"
)

var logger lura.Logger = nil

const logPrefix = "[PLUGIN:ACCESS-LOG]"

func SetLogger(l lura.Logger) {
	logger = l
}

func Debug(v ...interface{}) {
	logger.Debug(logPrefix, v)
}

func Info(v ...interface{}) {
	logger.Info(logPrefix, v)
}

func Warning(v ...interface{}) {
	logger.Warning(logPrefix, v)
}

func Error(v ...interface{}) {
	logger.Error(logPrefix, v)
}

func Critical(v ...interface{}) {
	logger.Critical(logPrefix, v)
}

func Fatal(v ...interface{}) {
	logger.Fatal(logPrefix, v)
}
