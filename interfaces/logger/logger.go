package logger

import (
	"github.com/sirupsen/logrus"
)

func LogInfo(logType string, data ...interface{}) {
	logrus.WithFields(logrus.Fields{"contentType": logType}).Info(data...)
}

func LogError(logType string, data ...interface{}) {
	logrus.WithFields(logrus.Fields{"contentType": logType}).Error(data...)
}
