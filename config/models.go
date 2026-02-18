package config

import "time"

type Config struct {
	ServerAddress    string
	DatabaseDSN      string
	CORSAllowOrigin  string
	CORSAllowMethods string
	CORSAllowHeaders string
	LogLevel         string
	LogFormat        string
	AuthJWTSecret    string
	AuthJWTIssuer    string
	AuthJWTAudience  string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	RefreshCookie    string
	RefreshPath      string
	RefreshSecure    bool
	RefreshSameSite  string
}

type CORSConfig struct {
	AllowOrigin  string
	AllowMethods string
	AllowHeaders string
}
