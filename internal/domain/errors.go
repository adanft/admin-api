package domain

import "errors"

const (
	InternalServerErrorMessage = "internal server error"
	NotFoundMessage            = "not found"
	ConflictMessage            = "conflict"
	BadRequestMessage          = "bad request"
	UnauthorizedMessage        = "unauthorized"

	UsernameExistsMessage     = ConflictMessage
	EmailExistsMessage        = ConflictMessage
	InvalidCredentialsMessage = UnauthorizedMessage
	InvalidEmailMessage       = BadRequestMessage
	WeakPasswordMessage       = BadRequestMessage
)

var (
	ErrInternalServerError = errors.New(InternalServerErrorMessage)
	ErrNotFound            = errors.New(NotFoundMessage)
	ErrConflict            = errors.New(ConflictMessage)
	ErrBadRequest          = errors.New(BadRequestMessage)
	ErrUnauthorized        = errors.New(UnauthorizedMessage)
	ErrUsernameExists      = errors.New(UsernameExistsMessage)
	ErrEmailExists         = errors.New(EmailExistsMessage)
	ErrInvalidCredentials  = errors.New(InvalidCredentialsMessage)
	ErrInvalidEmail        = errors.New(InvalidEmailMessage)
	ErrWeakPassword        = errors.New(WeakPasswordMessage)
)
