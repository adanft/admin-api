package token

import "errors"

const (
	InvalidTokenMessage      = "invalid token"
	ExpiredTokenMessage      = "token expired"
	InvalidConfiguration     = "invalid configuration"
	JWTSecretRequiredMessage = InvalidConfiguration
	JWTIssuerRequiredMessage = InvalidConfiguration
	JWTAudienceNeededMessage = InvalidConfiguration
	JWTAccessTTLMessage      = InvalidConfiguration
)

var (
	ErrInvalidToken      = errors.New(InvalidTokenMessage)
	ErrExpiredToken      = errors.New(ExpiredTokenMessage)
	ErrJWTSecretRequired = errors.New(JWTSecretRequiredMessage)
	ErrJWTIssuerRequired = errors.New(JWTIssuerRequiredMessage)
	ErrJWTAudienceNeeded = errors.New(JWTAudienceNeededMessage)
	ErrJWTAccessTTL      = errors.New(JWTAccessTTLMessage)
)
