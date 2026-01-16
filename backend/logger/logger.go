package logger

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var currentLevel LogLevel = INFO

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	level = strings.ToUpper(strings.TrimSpace(level))
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// SetLevel sets the current log level
func SetLevel(level LogLevel) {
	currentLevel = level
}

// GetLevel returns the current log level
func GetLevel() LogLevel {
	return currentLevel
}

// Init initializes the logger with a log level from environment variable
// Defaults to INFO if LOG_LEVEL is not set or invalid
func Init() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "INFO"
	}
	currentLevel = ParseLogLevel(levelStr)
	log.Printf("[INFO] Logger initialized with level: %s", currentLevel.String())
}

// shouldLog returns true if the given level should be logged
func shouldLog(level LogLevel) bool {
	return level >= currentLevel
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if shouldLog(DEBUG) {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if shouldLog(INFO) {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if shouldLog(WARN) {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if shouldLog(ERROR) {
		log.Printf("[ERROR] "+format, v...)
	}
}

// Fatal logs a fatal error and exits
func Fatal(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}
