package auth

import (
	"errors"
	"net/http"

	"admin.com/admin-api/internal/domain"
	httpErrors "admin.com/admin-api/internal/http/errors"
	appLogger "admin.com/admin-api/pkg/logger"
)

func writeAuthBusinessError(w http.ResponseWriter, r *http.Request, err error) {
	httpErrors.WriteBusinessError(w, r, err, appLogger.MsgAuthRequestFailed, mapAuthBusinessError)
}

func mapAuthBusinessError(err error) httpErrors.BusinessErrorMapping {
	mapped, ok := httpErrors.MapCommonBusinessError(err)
	if ok {
		return mapped
	}

	switch {
	case errors.Is(err, domain.ErrWeakPassword):
		return httpErrors.WeakPassword
	case errors.Is(err, domain.ErrInvalidCredentials):
		return httpErrors.InvalidCredentials
	case errors.Is(err, domain.ErrUnauthorized):
		return httpErrors.Unauthorized
	default:
		return httpErrors.Internal
	}
}
