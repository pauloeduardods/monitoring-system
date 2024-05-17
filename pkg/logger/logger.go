package logger

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
	Warning(format string, v ...interface{})
}

type logger struct{}

func NewLogger() Logger {
	return &logger{}
}

func (l *logger) Info(format string, v ...interface{}) {
	l.logWithCallerInfo("INFO", format, v...)
}

func (l *logger) Error(format string, v ...interface{}) {
	l.logWithCallerInfo("ERROR", format, v...)
}

func (l *logger) Warning(format string, v ...interface{}) {
	l.logWithCallerInfo("WARNING", format, v...)
}

func (l *logger) logWithCallerInfo(level, format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		file = file[strings.LastIndex(file, "/")+1:]
	}
	formattedMessage := fmt.Sprintf(format, v...)
	logMsg := fmt.Sprintf("%s:%d: %s - %s\n", file, line, level, formattedMessage)
	log.Println(logMsg)
}
