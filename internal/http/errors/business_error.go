package errors

import (
	stderrs "errors"
	"log/slog"
	"net/http"

	"admin.com/admin-api/internal/domain"
	"admin.com/admin-api/internal/http/middleware"
	"admin.com/admin-api/internal/http/response"
)

type BusinessErrorMapping struct {
	Status  int
	Code    string
	Message string
}

type BusinessErrorMapper func(err error) BusinessErrorMapping

func MapCommonBusinessError(err error) (BusinessErrorMapping, bool) {
	switch {
	case stderrs.Is(err, domain.ErrBadRequest):
		return InvalidPayload, true
	case stderrs.Is(err, domain.ErrInvalidEmail):
		return InvalidEmail, true
	case stderrs.Is(err, domain.ErrUsernameExists):
		return UsernameExists, true
	case stderrs.Is(err, domain.ErrEmailExists):
		return EmailExists, true
	case stderrs.Is(err, domain.ErrInternalServerError):
		return Internal, true
	default:
		return BusinessErrorMapping{}, false
	}
}

func WriteBusinessError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	logMessage string,
	mapper BusinessErrorMapper,
) {
	mapped := mapper(err)
	if mapped.Status == 0 {
		mapped.Status = http.StatusInternalServerError
	}

	if mapped.Status >= http.StatusInternalServerError {
		slog.Error(logMessage,
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"error_code", mapped.Code,
			"error", err,
		)
	}

	response.WriteErrorWithCode(w, mapped.Status, mapped.Code, mapped.Message)
}
