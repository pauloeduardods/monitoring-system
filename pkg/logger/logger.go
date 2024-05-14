package logger

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

type Logger struct{}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.logWithCallerInfo("INFO", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.logWithCallerInfo("ERROR", format, v...)
}

func (l *Logger) Warning(format string, v ...interface{}) {
	l.logWithCallerInfo("WARNING", format, v...)
}

func (l *Logger) logWithCallerInfo(level, format string, v ...interface{}) {
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
