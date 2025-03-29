package logger

import (
	"log"
	"os"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type StdLogger struct {
	level LogLevel
}

func NewStdLogger(level LogLevel) *StdLogger {
	log.SetOutput(os.Stdout)
	return &StdLogger{
		level: level,
	}
}

func (l *StdLogger) Info(format string, args ...any) {
	if l.level <= InfoLevel {
		log.Printf("[INFO] "+format, args...)
	}
}

func (l *StdLogger) Error(format string, args ...any) {
	if l.level <= ErrorLevel {
		log.Printf("[ERROR] "+format, args...)
	}
}

func (l *StdLogger) Debug(format string, args ...any) {
	if l.level <= DebugLevel {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func (l *StdLogger) Warn(format string, args ...any) {
	if l.level <= WarnLevel {
		log.Printf("[WARN] "+format, args...)
	}
}
