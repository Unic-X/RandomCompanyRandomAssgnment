package middleware

import (
	"net/http"
	"github.com/Unic-X/slow-server/metrics"
	"time"

	"github.com/google/uuid"
	"github.com/charmbracelet/log"
)

// LoggingMiddleware logs request information and timing
func ApplyLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}

		// Log request start
		startTime := time.Now()
		method := r.Method
		path := r.URL.Path
		log.Infof("[%s] Started %s %s", requestID, method, path)

		// Create a custom response writer to capture status code
		lrw := newLoggingResponseWriter(w)

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log request completion
		duration := time.Since(startTime)
		statusCode := lrw.statusCode
		log.Infof("[%s] Completed %s %s %d %s in %v", 
			requestID, method, path, statusCode, http.StatusText(statusCode), duration)
	})
}

// MetricsMiddleware collects metrics about requests
func ApplyMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		// Create a custom response writer to capture status code
		mrw := newLoggingResponseWriter(w)
		
		// Call the next handler
		next.ServeHTTP(mrw, r)
		
		// Record metrics
		duration := time.Since(startTime)
		path := r.URL.Path
		method := r.Method
		statusCode := mrw.statusCode
		
		// Update request counters
		metrics.RequestsTotal.WithLabelValues(path, method, string(statusCode)).Inc()
		
		// Update request duration histogram
		metrics.RequestDuration.WithLabelValues(path, method).Observe(float64(duration.Milliseconds()))
		
		// Track error rates
		if statusCode >= 400 {
			metrics.RequestErrors.WithLabelValues(path, method, string(statusCode)).Inc()
		}
	})
}

// loggingResponseWriter is a custom response writer that captures status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingResponseWriter creates a new logging response writer
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
