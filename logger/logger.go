package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// Logger struct with mutex for thread safety
type Logger struct {
	mu     sync.Mutex
	logger *log.Logger
}

// NewLogger initializes a logger
func NewLogger(output *os.File) *Logger {
	return &Logger{
		logger: log.New(output, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

var (
	infoLogger  = NewLogger(os.Stdout)
	debugLogger = NewLogger(os.Stdout)
	errorLogger = NewLogger(os.Stderr)
)

// LogInfo logs informational messages
func LogInfo(msg string, args ...interface{}) {
	infoLogger.mu.Lock()
	defer infoLogger.mu.Unlock()
	infoLogger.logger.Printf("[INFO] %s", fmt.Sprintf(msg, args...))
}

// LogDebug logs debug messages
func LogDebug(msg string, args ...interface{}) {
	debugLogger.mu.Lock()
	defer debugLogger.mu.Unlock()
	debugLogger.logger.Printf("[DEBUG] %s", fmt.Sprintf(msg, args...))
}

// LogError logs error messages
func LogError(msg string, args ...interface{}) {
	errorLogger.mu.Lock()
	defer errorLogger.mu.Unlock()
	errorLogger.logger.Printf("[ERROR] %s", fmt.Sprintf(msg, args...))
}
