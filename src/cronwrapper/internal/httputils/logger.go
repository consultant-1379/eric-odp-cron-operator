package httputils

import (
	"net/http"
	"runtime/debug"
	"time"

	log "eric-odp-cronwrapper/internal/logger"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// LoggingHandler logs the incoming HTTP request & its duration.
func LoggingHandler(serviceName string) func(http.Handler) http.Handler {
	mylog := log.WithFields(log.Fields{"serviceName": serviceName})

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					mylog.WithFields(log.Fields{"err": err, "trace": debug.Stack()}).
						Debug("http request")
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			mylog.WithFields(log.Fields{
				"status":   wrapped.status,
				"method":   r.Method,
				"path":     r.URL.EscapedPath(),
				"duration": time.Since(start),
			}).Debug("http response")
		}

		return http.HandlerFunc(fn)
	}
}
