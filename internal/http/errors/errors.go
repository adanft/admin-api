package errors

import (
	"net/http"

	"admin.com/admin-api/internal/domain"
)

var (
	InvalidID          = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "INVALID_ID", Message: domain.BadRequestMessage}
	InvalidPayload     = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "INVALID_PAYLOAD", Message: domain.BadRequestMessage}
	InvalidEmail       = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "INVALID_EMAIL", Message: domain.InvalidEmailMessage}
	WeakPassword       = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "WEAK_PASSWORD", Message: domain.WeakPasswordMessage}
	InvalidCredentials = BusinessErrorMapping{Status: http.StatusUnauthorized, Code: "INVALID_CREDENTIALS", Message: domain.InvalidCredentialsMessage}
	Unauthorized       = BusinessErrorMapping{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: domain.UnauthorizedMessage}
	UsernameExists     = BusinessErrorMapping{Status: http.StatusConflict, Code: "USERNAME_EXISTS", Message: domain.UsernameExistsMessage}
	EmailExists        = BusinessErrorMapping{Status: http.StatusConflict, Code: "EMAIL_EXISTS", Message: domain.EmailExistsMessage}
	AlreadyExists      = BusinessErrorMapping{Status: http.StatusConflict, Code: "ALREADY_EXISTS", Message: domain.ConflictMessage}
	NotFound           = BusinessErrorMapping{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: domain.NotFoundMessage}
	Internal           = BusinessErrorMapping{Status: http.StatusInternalServerError, Code: "INTERNAL", Message: domain.InternalServerErrorMessage}
	InvalidBody        = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: domain.BadRequestMessage}
	MalformedJSON      = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "MALFORMED_JSON", Message: domain.BadRequestMessage}
	InvalidFieldType   = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "INVALID_FIELD_TYPE", Message: domain.BadRequestMessage}
	BodyTooLarge       = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "BODY_TOO_LARGE", Message: domain.BadRequestMessage}
	InvalidContentType = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "INVALID_CONTENT_TYPE", Message: domain.BadRequestMessage}
	MultipleJSON       = BusinessErrorMapping{Status: http.StatusBadRequest, Code: "MULTIPLE_JSON_OBJECTS", Message: domain.BadRequestMessage}
)
