package logger

import (
	"log/syslog"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetSyslogLevelFromLogrusLevel(t *testing.T) {
	tests := []struct {
		name        string
		logrusLevel logrus.Level
		syslogLevel syslog.Priority
	}{
		{
			name:        "test info",
			logrusLevel: logrus.InfoLevel,
			syslogLevel: syslog.LOG_INFO,
		},
		{
			name:        "test debug",
			logrusLevel: logrus.DebugLevel,
			syslogLevel: syslog.LOG_DEBUG,
		},
		{
			name:        "test warn",
			logrusLevel: logrus.WarnLevel,
			syslogLevel: syslog.LOG_WARNING,
		},
		{
			name:        "test error",
			logrusLevel: logrus.ErrorLevel,
			syslogLevel: syslog.LOG_ERR,
		},
		{
			name:        "test panic",
			logrusLevel: logrus.PanicLevel,
			syslogLevel: syslog.LOG_ERR,
		},
		{
			name:        "test fatal",
			logrusLevel: logrus.FatalLevel,
			syslogLevel: syslog.LOG_ERR,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getSyslogLevelFromLogrusLevel(test.logrusLevel)
			if got != test.syslogLevel {
				t.Errorf("got %v, want %v", got, test.syslogLevel)
			}
		})
	}
}
