package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

var levelNames = map[LogLevel]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
	LevelPanic: "PANIC",
}

func (l LogLevel) String() string {
	if name, exists := levelNames[l]; exists {
		return name
	}
	return "UNKNOWN"
}

type Fields map[string]interface{}

type Logger struct {
	level      LogLevel
	logger     *log.Logger
	mu         sync.Mutex
	fields     Fields
	output     io.Writer
	timeFormat string
}

// New creates a new logger instance with the specified output destination
func New(output io.Writer) *Logger {
	return &Logger{
		level:      LevelInfo,
		logger:     log.New(output, "", log.LstdFlags),
		output:     output,
		timeFormat: "2006-01-02 15:04:05",
		fields:     make(Fields),
	}
}

// SetLevel sets the logging level for the logger
func (l *Logger) SetLevel(level LogLevel) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	return l
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(output io.Writer) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
	l.logger.SetOutput(output)
	return l
}

// SetTimeFormat sets the time format for the logger.
func (l *Logger) SetTimeFormat(format string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timeFormat = format
	return l
}

// AddField adds a field to the logger.
func (l *Logger) AddField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newLogger := l.clone()
	newLogger.fields[key] = value
	return newLogger
}

// AddFields adds multiple fields to the logger.
func (l *Logger) AddFields(fields Fields) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newLogger := l.clone()
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

func (l *Logger) clone() *Logger {
	fields := make(Fields)
	for k, v := range l.fields {
		fields[k] = v
	}
	return &Logger{
		level:      l.level,
		logger:     l.logger,
		output:     l.output,
		timeFormat: l.timeFormat,
		fields:     fields,
	}
}

func (l *Logger) log(level LogLevel, msg string) {
	if level < l.level {
		return
	}

	// Format the log entry
	entry := Fields{
		"level":   level.String(),
		"message": msg,
		"time":    time.Now().Format(l.timeFormat),
	}

	// Add context fields
	for k, v := range l.fields {
		entry[k] = v
	}

	// Marshal the log into a json using json.Marshal
	jsonData, err := json.Marshal(entry)
	if err != nil {
		logMsg := fmt.Sprintf(`{"level": "ERROR","message": "Failed to marshal log entry","error": "%s","time": "%s"}`,
			err.Error(),
			time.Now().Format(l.timeFormat))
		fmt.Fprintln(l.output, logMsg)
		return
	}

	fmt.Fprintln(l.output, string(jsonData))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(LevelError, msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.log(LevelFatal, msg)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(msg string) {
	l.log(LevelPanic, msg)
	panic(msg)
}

// Default logger instance
var defaultLogger = New(os.Stdout)

// Package-level functions for convenience
func SetLevel(level LogLevel) *Logger                { return defaultLogger.SetLevel(level) }
func SetOutput(w io.Writer) *Logger                  { return defaultLogger.SetOutput(w) }
func SetTimeFormat(format string) *Logger            { return defaultLogger.SetTimeFormat(format) }
func AddField(key string, value interface{}) *Logger { return defaultLogger.AddField(key, value) }
func AddFields(fields Fields) *Logger                { return defaultLogger.AddFields(fields) }
func Debug(msg string)                               { defaultLogger.Debug(msg) }
func Info(msg string)                                { defaultLogger.Info(msg) }
func Warn(msg string)                                { defaultLogger.Warn(msg) }
func Error(msg string)                               { defaultLogger.Error(msg) }
func Fatal(msg string)                               { defaultLogger.Fatal(msg) }
func Panic(msg string)                               { defaultLogger.Panic(msg) }
