package clog

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type resultType string
type color string

// LogLevel represents the logging level
type LogLevel int

const (
	// LevelError only shows errors
	LevelError LogLevel = iota
	// LevelWarning shows errors and warnings
	LevelWarning
	// LevelInfo shows all messages
	LevelInfo
)

// Logger is the main logger instance
type Logger struct {
	Writer    io.Writer
	Level     LogLevel
	ColorMode bool
}

// DefaultLogger is the package-level logger with default settings
var DefaultLogger = &Logger{
	Writer:    os.Stdout,
	Level:     LevelInfo,
	ColorMode: true,
}

// CheckResult is the result of dependency checking
type CheckResult struct {
	Type       resultType
	Message    string
	Color      color
	IndentSize int
}

// NewLogger creates a new logger with the given options
func NewLogger(writer io.Writer, level LogLevel) *Logger {
	return &Logger{
		Writer:    writer,
		Level:     level,
		ColorMode: true,
	}
}

// SetLogLevel updates the log level for the default logger
func SetLogLevel(level LogLevel) {
	DefaultLogger.Level = level
}

// DisableColor turns off colored output
func DisableColor() {
	DefaultLogger.ColorMode = false
}

// EnableColor turns on colored output
func EnableColor() {
	DefaultLogger.ColorMode = true
}

// Error logs an error message
func Error(msg string) {
	DefaultLogger.Error(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	if l.Level >= LevelError {
		res := NewError(msg)
		l.PrintMessage(res)
	}
}

// Warning logs a warning message
func Warning(msg string) {
	DefaultLogger.Warning(msg)
}

// Warning logs a warning message
func (l *Logger) Warning(msg string) {
	if l.Level >= LevelWarning {
		res := NewWarning(msg)
		l.PrintMessage(res)
	}
}

// Info logs an info message
func Info(msg string) {
	DefaultLogger.Info(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	if l.Level >= LevelInfo {
		res := NewInfo(msg)
		l.PrintMessage(res)
	}
}

// NewError creates a new error result
func NewError(msg string) CheckResult {
	return CheckResult{
		Type:       ResultErr,
		Message:    msg,
		Color:      Red,
		IndentSize: 0,
	}
}

// NewInfo creates a new info result
func NewInfo(msg string) CheckResult {
	return CheckResult{
		Type:       ResultInfo,
		Message:    msg,
		Color:      Teal,
		IndentSize: 0,
	}
}

// NewWarning creates a new warning result
func NewWarning(msg string) CheckResult {
	return CheckResult{
		Type:       ResultWarning,
		Message:    msg,
		Color:      Yellow,
		IndentSize: 0,
	}
}

// WithIndent creates a copy of the check result with the specified indentation
func (cr CheckResult) WithIndent(indentSize int) CheckResult {
	cr.IndentSize = indentSize
	return cr
}

// PrintColorMessage prints a message with color formatting
func PrintColorMessage(cr CheckResult) {
	DefaultLogger.PrintMessage(cr)
}

// PrintMessage prints a message with the logger's settings
func (l *Logger) PrintMessage(cr CheckResult) {
	indent := ""
	if cr.IndentSize > 0 {
		indent = strings.Repeat("  ", cr.IndentSize)
	}

	if l.ColorMode {
		fmt.Fprintf(l.Writer, "%s%s%-11s%s\n%s", indent, cr.Color, "["+cr.Type+"]", cr.Message, Reset)
	} else {
		fmt.Fprintf(l.Writer, "%s%-11s%s\n", indent, "["+cr.Type+"]", cr.Message)
	}
}
