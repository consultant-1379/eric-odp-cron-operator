package logger

import (
	"io"
	"log/syslog"
	"net"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

// AdpLogJSONFormatter The fields format according to RFC5423.
var adpLogJSONFormatter = logrus.JSONFormatter{
	TimestampFormat: adpLogTimeFormat,
	FieldMap: logrus.FieldMap{
		logrus.FieldKeyTime:  "timestamp",
		logrus.FieldKeyLevel: "severity",
		logrus.FieldKeyMsg:   "message",
	},
}

type logrusEntry struct {
	entry *logrus.Entry
}

type logrusLogger struct {
	logger *logrus.Logger
}

func (l *logrusLogger) Debug(msg ...interface{}) {
	l.logger.Debug(msg...)
}

func (l *logrusLogger) Warn(msg ...interface{}) {
	l.logger.Warn(msg...)
}

func (l *logrusLogger) Fatal(msg ...interface{}) {
	l.logger.Fatal(msg...)
}

func (l *logrusLogger) Panic(msg ...interface{}) {
	l.logger.Panic(msg...)
}

func (l *logrusLogger) Error(msg ...interface{}) {
	l.logger.Error(msg...)
}

func (l *logrusLogger) Info(msg ...interface{}) {
	l.logger.Info(msg...)
}

func (l *logrusLogger) WithFields(fields Fields) Logger {
	return &logrusEntry{
		entry: l.logger.WithFields(convertToLogrusFields(fields)),
	}
}

func (l *logrusLogger) SetLevel(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return &LogError{"set level error", err}
	}
	l.logger.SetLevel(logLevel)

	return nil
}

func (l *logrusEntry) Debug(msg ...interface{}) {
	l.entry.Debug(msg...)
}

func (l *logrusEntry) Warn(msg ...interface{}) {
	l.entry.Warn(msg...)
}

func (l *logrusEntry) Error(msg ...interface{}) {
	l.entry.Error(msg...)
}

func (l *logrusEntry) Info(msg ...interface{}) {
	l.entry.Info(msg...)
}

func (l *logrusEntry) Fatal(msg ...interface{}) {
	l.entry.Fatal(msg...)
}

func (l *logrusEntry) Panic(msg ...interface{}) {
	l.entry.Panic(msg...)
}

func (l *logrusEntry) WithFields(fields Fields) Logger {
	return &logrusEntry{
		entry: l.entry.WithFields(convertToLogrusFields(fields)),
	}
}

func (l *logrusEntry) SetLevel(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return &LogError{"set level error", err}
	}
	l.entry.Logger.SetLevel(logLevel)

	return nil
}

// newLogrus The method creates a simple logrus logger conforming to
// adp logging schema.
func newLogrus(logConfig *LogConfig) (Logger, error) {
	logrusLog := logrus.New()

	logrusLog.SetFormatter(&adpLogJSONFormatter)

	logLevel, err := logrus.ParseLevel(logConfig.LogLevel)
	if err != nil {
		return nil, &LogError{"new logrus create error", err}
	}
	logrusLog.SetLevel(logLevel)

	if logConfig.LogStreamingMethod == "direct" || logConfig.LogStreamingMethod == "dual" {
		if syslogHook := buildSyslogHook(logConfig, logLevel); syslogHook != nil {
			logrusLog.AddHook(syslogHook)
		} else {
			logrusLog.Warning("Unable to set syslog hook, logs won't be streamed to log transformer")
		}
	}

	if logConfig.LogStreamingMethod == "direct" {
		logrusLog.Out = io.Discard
	}

	// Add the keyValues if defined in the Ctx to the logger
	// Adp requires service_id, version_id
	if logConfig.Fields != nil {
		logrusLogEntry := logrusLog.WithFields(convertToLogrusFields(logConfig.Fields))

		return &logrusEntry{entry: logrusLogEntry}, nil
	}

	return &logrusLogger{logger: logrusLog}, nil
}

func convertToLogrusFields(fields map[string]interface{}) logrus.Fields {
	logrusFields := logrus.Fields{}
	for index, val := range fields {
		logrusFields[index] = val
	}

	return logrusFields
}

func buildSyslogHook(logConfig *LogConfig, logLevel logrus.Level) *lSyslog.SyslogHook {
	if logConfig.LogstashHost == "" || logConfig.LogstashSyslogPort == "" {
		return nil
	}
	logStashHostPort := net.JoinHostPort(logConfig.LogstashHost, logConfig.LogstashSyslogPort)
	syslogLogLevel := getSyslogLevelFromLogrusLevel(logLevel)

	syslogHook, err := lSyslog.NewSyslogHook("tcp", logStashHostPort, syslogLogLevel, "")
	if err != nil {
		return nil
	}

	return syslogHook
}

func getSyslogLevelFromLogrusLevel(logLevel logrus.Level) syslog.Priority {
	switch logLevel.String() {
	case InfoLevel:
		return syslog.LOG_INFO
	case DebugLevel:
		return syslog.LOG_DEBUG
	case WarnLevel:
		return syslog.LOG_WARNING
	case ErrorLevel, PanicLevel, FatalLevel:
		return syslog.LOG_ERR
	}

	return syslog.LOG_INFO
}
