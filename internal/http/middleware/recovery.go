package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"admin.com/admin-api/internal/domain"
	"admin.com/admin-api/internal/http/response"
	appLogger "admin.com/admin-api/pkg/logger"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := newStatusRecorder(w)

		defer func() {
			if rec := recover(); rec != nil {
				requestID := RequestIDFromContext(r.Context())
				if requestID == "" {
					requestID = recorder.Header().Get(RequestIDHeader)
				}

				slog.Error(appLogger.MsgPanicRecovered,
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"client_ip", clientIP(r),
					"panic", rec,
					"stack_trace", string(debug.Stack()),
				)

				if !recorder.written {
					response.WriteError(recorder, http.StatusInternalServerError, domain.InternalServerErrorMessage)
				}
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}
