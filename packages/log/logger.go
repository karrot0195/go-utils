package log

import (
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger() (*Logger, error) {
	l := &Logger{
		logrus.New(),
	}

	return l, nil
}
