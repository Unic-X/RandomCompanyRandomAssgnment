package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Unic-X/slow-server/logger"
	"github.com/Unic-X/slow-server/metrics"
	"github.com/vedadiyan/lokiclient"

	"github.com/google/uuid"
)

// LoggingMiddleware logs request information and timing
func ApplyLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}

		startTime := time.Now()
		method := r.Method
		path := r.URL.Path

		lrw := newLoggingResponseWriter(w)

		next.ServeHTTP(lrw, r)

		duration := time.Since(startTime)
		statusCode := lrw.statusCode

        customStream := lokiclient.NewStreamCustom(map[string]string{
            "requestID": requestID, 
            "method": method, 
            "path": path, 
            "status": string(statusCode), 
            "statusText": http.StatusText(statusCode), 
            "duration":duration.String(),
        })
		log.Info(context.Background(),customStream,"Info Related to this")
    })
}

func ApplyMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log := logger.GetLogger()
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
            stream := lokiclient.NewStream("slow-server", "middleware", "ApplyMetricsMiddleware", "trace123")
            log.Info(context.Background(), stream, "Status Code 400 :<")
			metrics.RequestErrors.WithLabelValues(path, method, string(statusCode)).Inc()
		}
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
