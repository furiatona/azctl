package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// Level represents the logging level
type Level string

const (
	// LevelDebug represents debug level logging
	LevelDebug Level = "debug"
	// LevelInfo represents info level logging
	LevelInfo Level = "info"
	// LevelWarn represents warning level logging
	LevelWarn Level = "warn"
	// LevelError represents error level logging
	LevelError Level = "error"
	// LevelFatal represents fatal level logging
	LevelFatal Level = "fatal"
)

// Formatter represents the logging format
type Formatter string

const (
	// FormatterText represents text format
	FormatterText Formatter = "text"
	// FormatterJSON represents JSON format
	FormatterJSON Formatter = "json"
)

// Config holds logging configuration
type Config struct {
	Level     Level     `json:"level" yaml:"level"`
	Formatter Formatter `json:"formatter" yaml:"formatter"`
	Output    string    `json:"output" yaml:"output"`
	File      string    `json:"file" yaml:"file"`
}

// Logger is the main logging interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// logger implements the Logger interface
type logger struct {
	logrus *logrus.Logger
}

// New creates a new logger with the given configuration
func New(config Config) (Logger, error) {
	l := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	l.SetLevel(level)

	// Set formatter
	switch config.Formatter {
	case FormatterJSON:
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return "", fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	case FormatterText:
		fallthrough
	default:
		l.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := filepath.Base(f.File)
				return "", fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
	}

	// Set output
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		l.SetOutput(io.MultiWriter(os.Stdout, file))
	} else if config.Output != "" {
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		l.SetOutput(file)
	} else {
		l.SetOutput(os.Stdout)
	}

	// Enable caller info for better debugging
	l.SetReportCaller(true)

	return &logger{logrus: l}, nil
}

// Debug logs a debug message
func (l *logger) Debug(args ...interface{}) {
	l.logrus.Debug(args...)
}

// Debugf logs a formatted debug message
func (l *logger) Debugf(format string, args ...interface{}) {
	l.logrus.Debugf(format, args...)
}

// Info logs an info message
func (l *logger) Info(args ...interface{}) {
	l.logrus.Info(args...)
}

// Infof logs a formatted info message
func (l *logger) Infof(format string, args ...interface{}) {
	l.logrus.Infof(format, args...)
}

// Warn logs a warning message
func (l *logger) Warn(args ...interface{}) {
	l.logrus.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *logger) Warnf(format string, args ...interface{}) {
	l.logrus.Warnf(format, args...)
}

// Error logs an error message
func (l *logger) Error(args ...interface{}) {
	l.logrus.Error(args...)
}

// Errorf logs a formatted error message
func (l *logger) Errorf(format string, args ...interface{}) {
	l.logrus.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(args ...interface{}) {
	l.logrus.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.logrus.Fatalf(format, args...)
}

// WithField adds a field to the logger
func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{logrus: l.logrus.WithField(key, value).Logger}
}

// WithFields adds multiple fields to the logger
func (l *logger) WithFields(fields map[string]interface{}) Logger {
	return &logger{logrus: l.logrus.WithFields(logrus.Fields(fields)).Logger}
}

// Global logger instance
var globalLogger Logger

// Init initializes the global logger
func Init(config Config) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Get returns the global logger
func Get() Logger {
	if globalLogger == nil {
		// Initialize with default config if not initialized
		config := Config{
			Level:     LevelInfo,
			Formatter: FormatterText,
		}
		if err := Init(config); err != nil {
			panic(fmt.Sprintf("failed to initialize logger: %v", err))
		}
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(args ...interface{})                 { Get().Debug(args...) }
func Debugf(format string, args ...interface{}) { Get().Debugf(format, args...) }
func Info(args ...interface{})                  { Get().Info(args...) }
func Infof(format string, args ...interface{})  { Get().Infof(format, args...) }
func Warn(args ...interface{})                  { Get().Warn(args...) }
func Warnf(format string, args ...interface{})  { Get().Warnf(format, args...) }
func Error(args ...interface{})                 { Get().Error(args...) }
func Errorf(format string, args ...interface{}) { Get().Errorf(format, args...) }
func Fatal(args ...interface{})                 { Get().Fatal(args...) }
func Fatalf(format string, args ...interface{}) { Get().Fatalf(format, args...) }
