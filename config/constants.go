package config

import "time"

const (
	defaultAddress          = ":9090"
	defaultDatabaseSSLMode  = "disable"
	defaultCORSAllowOrigin  = "*"
	defaultCORSAllowMethods = "GET, POST, PUT, DELETE, OPTIONS"
	defaultCORSAllowHeaders = "Content-Type, Authorization"
	defaultLogLevel         = "info"
	defaultLogFormat        = "json"
	defaultAuthJWTSecret    = "change-me-dev-secret"
	defaultAuthJWTIssuer    = "admin-api"
	defaultAuthJWTAudience  = "admin-api-client"
	defaultAccessTokenTTL   = 15 * time.Minute
	defaultRefreshTokenTTL  = 7 * 24 * time.Hour
	defaultRefreshCookie    = "refresh_token"
	defaultRefreshPath      = "/auth"
	defaultRefreshSecure    = false
	defaultRefreshSameSite  = "Lax"
)
