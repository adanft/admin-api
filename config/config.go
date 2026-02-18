package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

func Load() (Config, error) {
	dbHost, err := getRequiredEnv("DATABASE_HOST")
	if err != nil {
		return Config{}, err
	}
	dbPort, err := getRequiredEnv("DATABASE_PORT")
	if err != nil {
		return Config{}, err
	}
	dbUser, err := getRequiredEnv("DATABASE_USER")
	if err != nil {
		return Config{}, err
	}
	dbPass, err := getRequiredEnv("DATABASE_PASS")
	if err != nil {
		return Config{}, err
	}
	dbName, err := getRequiredEnv("DATABASE_NAME")
	if err != nil {
		return Config{}, err
	}

	serverAddress := getEnvOrDefault("SERVER_ADDRESS", defaultAddress)
	sslMode := getEnvOrDefault("DATABASE_SSL_MODE", defaultDatabaseSSLMode)
	corsAllowOrigin := getEnvOrDefault("CORS_ALLOW_ORIGIN", defaultCORSAllowOrigin)
	corsAllowMethods := getEnvOrDefault("CORS_ALLOW_METHODS", defaultCORSAllowMethods)
	corsAllowHeaders := getEnvOrDefault("CORS_ALLOW_HEADERS", defaultCORSAllowHeaders)
	logLevel := getEnvOrDefault("LOG_LEVEL", defaultLogLevel)
	logFormat := getEnvOrDefault("LOG_FORMAT", defaultLogFormat)
	authJWTSecret := getEnvOrDefault("AUTH_JWT_SECRET", defaultAuthJWTSecret)
	authJWTIssuer := getEnvOrDefault("AUTH_JWT_ISSUER", defaultAuthJWTIssuer)
	authJWTAudience := getEnvOrDefault("AUTH_JWT_AUDIENCE", defaultAuthJWTAudience)
	accessTokenTTL, err := getDurationEnvOrDefault("AUTH_ACCESS_TOKEN_TTL", defaultAccessTokenTTL)
	if err != nil {
		return Config{}, err
	}
	refreshTokenTTL, err := getDurationEnvOrDefault("AUTH_REFRESH_TOKEN_TTL", defaultRefreshTokenTTL)
	if err != nil {
		return Config{}, err
	}
	refreshCookie := getEnvOrDefault("AUTH_REFRESH_COOKIE_NAME", defaultRefreshCookie)
	refreshPath := getEnvOrDefault("AUTH_REFRESH_COOKIE_PATH", defaultRefreshPath)
	refreshSecure, err := getBoolEnvOrDefault("AUTH_REFRESH_COOKIE_SECURE", defaultRefreshSecure)
	if err != nil {
		return Config{}, err
	}
	refreshSameSite := getEnvOrDefault("AUTH_REFRESH_COOKIE_SAMESITE", defaultRefreshSameSite)
	dsn := buildPostgresDSN(dbHost, dbPort, dbUser, dbPass, dbName, sslMode)

	return Config{
		ServerAddress:    serverAddress,
		DatabaseDSN:      dsn,
		CORSAllowOrigin:  corsAllowOrigin,
		CORSAllowMethods: corsAllowMethods,
		CORSAllowHeaders: corsAllowHeaders,
		LogLevel:         logLevel,
		LogFormat:        logFormat,
		AuthJWTSecret:    authJWTSecret,
		AuthJWTIssuer:    authJWTIssuer,
		AuthJWTAudience:  authJWTAudience,
		AccessTokenTTL:   accessTokenTTL,
		RefreshTokenTTL:  refreshTokenTTL,
		RefreshCookie:    refreshCookie,
		RefreshPath:      refreshPath,
		RefreshSecure:    refreshSecure,
		RefreshSameSite:  refreshSameSite,
	}, nil
}

func buildPostgresDSN(host string, port string, user string, pass string, dbName string, sslMode string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbName,
	}

	query := u.Query()
	query.Set("sslmode", sslMode)
	u.RawQuery = query.Encode()

	return u.String()
}

func getRequiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	return value, nil
}

func getEnvOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

func getDurationEnvOrDefault(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s has invalid duration: %w", name, err)
	}

	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}

	return duration, nil
}

func getBoolEnvOrDefault(name string, fallback bool) (bool, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s has invalid boolean value: %w", name, err)
	}

	return parsed, nil
}
