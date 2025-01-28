package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/logging"
)

type Logger struct {
	mu          sync.Mutex
	logger      *log.Logger
	prefix      string
	cloudLogger *logging.Logger
}

var (
	InfoLogger  *Logger
	DebugLogger *Logger
	ErrorLogger *Logger
)

// NewLogger initializes a logger
func NewLogger(output *os.File, cloudLogger *logging.Logger) *Logger {
	return &Logger{
		logger:      log.New(output, "", log.Ldate|log.Ltime|log.Lshortfile),
		cloudLogger: cloudLogger,
	}
}

// InitializeLoggers initializes loggers with Cloud Logging support
func InitializeLoggers(ctx context.Context, projectID string) error {
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create logging client: %w", err)
	}

	InfoLogger = NewLogger(os.Stdout, client.Logger("info"))
	DebugLogger = NewLogger(os.Stdout, client.Logger("debug"))
	ErrorLogger = NewLogger(os.Stderr, client.Logger("error"))
	return nil
}

// SetPrefix sets a custom prefix for the logger
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// ClearPrefix clears the prefix
func (l *Logger) ClearPrefix() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = ""
}

// getPrefix retrieves the prefix if set, or an empty string otherwise
func (l *Logger) getPrefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.prefix != "" {
		return fmt.Sprintf("[%s] ", l.prefix)
	}
	return ""
}

// Log writes a log entry to stdout and GCP Cloud Logging
func (l *Logger) Log(level, msg string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, args...)
	prefixedMsg := fmt.Sprintf("%s%s", l.getPrefix(), formattedMsg)

	// Log to stdout
	l.logger.Printf("[%s] %s", level, prefixedMsg)

	// Log to Cloud Logging
	if l.cloudLogger != nil {
		l.cloudLogger.Log(logging.Entry{
			Severity: logging.ParseSeverity(level),
			Payload:  prefixedMsg,
		})
	}
}

// LogInfo logs informational messages
func LogInfo(msg string, args ...interface{}) {
	InfoLogger.Log("INFO", msg, args...)
}

// LogDebug logs debug messages
func LogDebug(msg string, args ...interface{}) {
	DebugLogger.Log("DEBUG", msg, args...)
}

// LogError logs error messages
func LogError(msg string, args ...interface{}) {
	ErrorLogger.Log("ERROR", msg, args...)
}
