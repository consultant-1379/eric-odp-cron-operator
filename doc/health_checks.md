# Health checks

## Implementation details:
Liveness and readiness probes are configured in cmd/eric-odp-cron-operator/server.go initHealthCheck().
- Liveness probe can be fetched by "/health/liveness" endpoint and "9797" port by default. The default port can be changed by
HEALTH_CHECK_PORT environment variable. Change default implementation based on application requirements.
- Readiness probe can be fetched by "/health/readiness" endpoint and "9797" port by default. The default port can be changed by
HEALTH_CHECK_PORT environment variable. Change default implementation based on application requirements.