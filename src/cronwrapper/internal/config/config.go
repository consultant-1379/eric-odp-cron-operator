package config

import (
	"os"
	"strconv"
	"strings"
)

// Config The parameters for service configuration.
type Config struct {
	LocalPort       int
	MetricsPort     int
	HealthCheckPort int

	LogControlFile     string
	LogstashHost       string
	LogstashSyslogPort string
	LogStreamingMethod string

	FaultHandlingEnabled bool
	UseRESTForFI         bool

	OdpFactoryHost   string
	OdpFactoryPort   string
	OdpFactorySchema string

	RetrySleepTime int
	MaxSleepTime   int

	//Parameters set by the Cron Scheduler
	Username     string
	Command      string
	Hostname     string
	UniqueCronId string
}

var instance *Config

func GetConfig() *Config {
	if instance == nil {
		instance = &Config{
			LocalPort:       getOsEnvInt("LOCAL_PORT", defaultLocalPort),
			HealthCheckPort: getOsEnvInt("HEALTH_CHECK_PORT", defaultHealthCheckPort),
			MetricsPort:     getOsEnvInt("METRICS_PORT", defaultMetricsPort),

			// LogControlFile INT.LOG.CTRL for controlling log severity
			LogControlFile:     getOsEnvString("LOG_CTRL_FILE", ""),
			LogstashHost:       getOsEnvString("LOGSTASH_HOST", ""),
			LogstashSyslogPort: getOsEnvString("LOGSTASH_SYSLOG_PORT", defaultLogstashSyslogPort),
			LogStreamingMethod: getOsEnvString("LOG_STREAMING_METHOD", defaultLogStreamingMethod),

			OdpFactoryHost:   getOsEnvString("ODP_FACTORY_HOST", defaultOdpFactoryHost),
			OdpFactoryPort:   getOsEnvString("ODP_FACTORY_PORT", defaultOdpFactoryPort),
			OdpFactorySchema: getOsEnvString("ODP_FACTORY_SCHEMA", defaultOdpFactorySchema),

			RetrySleepTime: getOsEnvInt("SLEEP_TIME", defaultSleepTime),
			MaxSleepTime:   getOsEnvInt("MAX_SLEEP_TIME", defaultMaxSleepTime),

			//Read the environment variables set by the cron scheduler
			Username:     getOsEnvString("ODP-CRON-POD-USERNAME", ""),
			Command:      getOsEnvString("ODP-CRON-POD-CMD", ""),
			Hostname:     getOsEnvString("HOSTNAME", ""),
			UniqueCronId: getOsEnvString("UNIQUE_CRON_ID", ""),
		}
	}

	return instance
}

func getOsEnvInt(envName string, defaultValue int) (result int) {
	envValue := strings.TrimSpace(os.Getenv(envName))

	result, err := strconv.Atoi(envValue)
	if err != nil {
		result = defaultValue
	}

	return
}

func getOsEnvString(envName string, defaultValue string) (result string) {
	result = strings.TrimSpace(os.Getenv(envName))
	if result == "" {
		result = defaultValue
	}

	return
}
