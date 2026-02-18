package user

import (
	"errors"
	"net/http"

	"admin.com/admin-api/internal/domain"
	httpErrors "admin.com/admin-api/internal/http/errors"
	appLogger "admin.com/admin-api/pkg/logger"
)

func writeUserBusinessError(w http.ResponseWriter, r *http.Request, err error) {
	httpErrors.WriteBusinessError(w, r, err, appLogger.MsgUserRequestFailed, mapUserBusinessError)
}

func mapUserBusinessError(err error) httpErrors.BusinessErrorMapping {
	mapped, ok := httpErrors.MapCommonBusinessError(err)
	if ok {
		return mapped
	}

	switch {
	case errors.Is(err, domain.ErrConflict):
		return httpErrors.AlreadyExists
	case errors.Is(err, domain.ErrNotFound):
		return httpErrors.NotFound
	default:
		return httpErrors.Internal
	}
}
