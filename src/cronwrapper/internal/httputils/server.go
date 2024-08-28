package httputils

import (
	"net/http"
	"os"

	stdlog "log"
)

// ServerConfig A struct for holding common server traits.
type ServerConfig struct {
	Addr       string
	Handler    http.Handler
	ServerName string
}

func errorLogger(name string) *stdlog.Logger {
	logger := stdlog.New(os.Stdout, name, stdlog.LstdFlags)

	return logger
}

// NewServer The method returns a new server based on the
// incoming configuration.
func NewServer(config *ServerConfig) *http.Server {
	logmw := LoggingHandler(config.ServerName)

	return &http.Server{
		Addr:         config.Addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      logmw(config.Handler),
		ErrorLog:     errorLogger(config.ServerName),
	}
}
