package application

import (
	log "github.com/sirupsen/logrus"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger(logger *log.Logger) *Logger {
	return &Logger{logger: logger}
}

func (l *Logger) Error(message string) {
	l.logger.Error(message)
}

func (l *Logger) Info(message string) {
	l.logger.Info(message)
}
