package logger

import (
	"log"
	"os"
)

// Logger holds different loggers for various levels
type Logger struct {
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
}

// NewLogger initializes and returns a Logger instance with configured loggers
func NewLogger() *Logger {
	return &Logger{
		InfoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		WarningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLogger:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs informational messages
func (l *Logger) Info(msg string) {
	l.InfoLogger.Println(msg)
}

// Warning logs warning messages
func (l *Logger) Warning(msg string) {
	l.WarningLogger.Println(msg)
}

// Error logs error messages
func (l *Logger) Error(msg string) {
	l.ErrorLogger.Println(msg)
}
