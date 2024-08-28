// Package logger implements a simple logging abstraction. It defines a type Logger
// containing methods which are considered required by the ADP logging guidelines
// The package contains a predefined 'global' logger which together with the helper
// methods (Info/Error/Fatal/Debug/WithFields) can be used directly.
// The package also provides method to create a new instance, satisfying the interface
// which could be also be used
package logger

import (
	"fmt"
	"eric-odp-cronwrapper/internal/config"
	stdlog "log"
	"sync"
)

const (
	// adpLogTimeFormat The timestamp format according to RFC5424.
	adpLogTimeFormat = "2006-01-02T15:04:05.999Z07:00"
)

// Global logger instance.
var log Logger

var once sync.Once

const (
	FatalLevel = "fatal"
	PanicLevel = "panic"
	ErrorLevel = "error"
	WarnLevel  = "warning"
	InfoLevel  = "info"
	DebugLevel = "debug"
)

type LogError struct {
	Msg string
	Err error
}

func (e *LogError) Unwrap() error {
	return e.Err
}

func (e *LogError) Error() string {
	return fmt.Sprintf("%s: %s", e.Msg, e.Err.Error())
}

func init() {
	var err error
	appConfig := config.GetConfig()
	logConfig := LogConfig{
		LogLevel:           "debug",
		LogstashHost:       appConfig.LogstashHost,
		LogstashSyslogPort: appConfig.LogstashSyslogPort,
		LogStreamingMethod: appConfig.LogStreamingMethod,
	}
	log, err = NewLogger(&logConfig)
	if err != nil {
		stdlog.Panic("Unable to initialize logger")
	}
}

// Fields The key store of fields.
type Fields map[string]interface{}

// LogConfig The key store of configs.
type LogConfig struct {
	LogLevel           string
	LogstashHost       string
	LogstashSyslogPort string
	LogStreamingMethod string
	Fields
}

// Logger interface The following are methods are used.
type Logger interface {
	Info(...interface{})
	Error(...interface{})
	Warn(...interface{})
	Debug(...interface{})
	Fatal(...interface{})
	Panic(...interface{})
	WithFields(Fields) Logger
	SetLevel(string) error
}

// Debug The helper method for debug logs on default log instance.
func Debug(msg ...interface{}) {
	log.Debug(msg...)
}

// Info The helper method for info logs on default log instance.
func Info(msg ...interface{}) {
	log.Info(msg...)
}

// Error The helper method for error logs on default log instance.
func Error(msg ...interface{}) {
	log.Error(msg...)
}

// Warn The helper method for warn logs on default log instance.
func Warn(msg ...interface{}) {
	log.Warn(msg...)
}

// Fatal The helper method for fatal logs on default log instance.
func Fatal(msg ...interface{}) {
	log.Fatal(msg...)
}

// Panic The helper method for panic logs on default log instance.
func Panic(msg ...interface{}) {
	log.Panic(msg...)
}

// WithFields The helper method for debug logs on default log instance.
func WithFields(keyValues Fields) Logger {
	return log.WithFields(keyValues)
}

// SetLevel The method sets the log level of the default log instance.
func SetLevel(level string) error {
	err := log.SetLevel(level)
	if err != nil {
		return &LogError{"set level error", err}
	}

	return nil
}

// SetDefaultFields The method add fields to the default log instance.
func SetDefaultFields(fields Fields) {
	once.Do(
		func() {
			log = log.WithFields(fields)
		})
}

// NewLogger Creates a logger instance based on configuration.
func NewLogger(logConfig *LogConfig) (Logger, error) {
	logger, err := newLogrus(logConfig)
	if err != nil {
		return nil, err
	}

	return logger, nil
}
