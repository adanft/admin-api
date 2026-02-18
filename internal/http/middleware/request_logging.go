package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	appLogger "admin.com/admin-api/pkg/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.written = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if !r.written {
		r.written = true
	}
	size, err := r.ResponseWriter.Write(body)
	r.size += size
	return size, err
}

func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := newStatusRecorder(w)

		defer func() {
			requestID := RequestIDFromContext(r.Context())
			if requestID == "" {
				requestID = recorder.Header().Get(RequestIDHeader)
			}

			logArgs := []any{
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"response_bytes", recorder.size,
				"client_ip", clientIP(r),
			}

			switch {
			case recorder.status >= http.StatusInternalServerError:
				slog.Error(appLogger.MsgHTTPRequest, logArgs...)
			case recorder.status >= http.StatusBadRequest:
				slog.Warn(appLogger.MsgHTTPRequest, logArgs...)
			default:
				slog.Info(appLogger.MsgHTTPRequest, logArgs...)
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}

func clientIP(r *http.Request) string {
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		if idx := strings.Index(forwardedFor, ","); idx != -1 {
			return strings.TrimSpace(forwardedFor[:idx])
		}
		return forwardedFor
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-Ip")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
