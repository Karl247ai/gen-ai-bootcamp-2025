package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger struct {
	output io.Writer
	level  Level
}

type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Fields    interface{} `json:"fields,omitempty"`
}

func New(output io.Writer, level Level) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		output: output,
		level:  level,
	}
}

func (l *Logger) log(level Level, msg string, fields interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     levelToString(level),
		Message:   msg,
		Fields:    fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling log entry: %v\n", err)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(DebugLevel, msg, combineFields(fields...))
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(InfoLevel, msg, combineFields(fields...))
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.log(WarnLevel, msg, combineFields(fields...))
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.log(ErrorLevel, msg, combineFields(fields...))
}

// WithContext returns a context with the logger attached
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext retrieves the logger from the context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*Logger); ok {
		return logger
	}
	return New(os.Stdout, InfoLevel)
}

type loggerKey struct{}

func levelToString(level Level) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func combineFields(fields ...interface{}) interface{} {
	if len(fields) == 0 {
		return nil
	}
	if len(fields) == 1 {
		return fields[0]
	}
	return fields
} 