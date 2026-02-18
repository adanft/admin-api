package app

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"admin.com/admin-api/config"
	httpcookie "admin.com/admin-api/internal/http/cookie"
	authhttp "admin.com/admin-api/internal/http/handler/auth"
	userhttp "admin.com/admin-api/internal/http/handler/user"
	"admin.com/admin-api/internal/http/middleware"
	authrepo "admin.com/admin-api/internal/repository/postgres/auth"
	userrepo "admin.com/admin-api/internal/repository/postgres/user"
	securitytoken "admin.com/admin-api/internal/security/token"
	authapp "admin.com/admin-api/internal/usecase/auth"
	userapp "admin.com/admin-api/internal/usecase/user"
	"admin.com/admin-api/pkg/crypto"
	"github.com/uptrace/bun"
)

func NewHandler(appCfg config.Config, dbConn *bun.DB) (http.Handler, error) {
	userStore := userrepo.NewUserRepository(dbConn)
	userUseCase := userapp.NewUserUseCase(userStore, crypto.HashPassword)

	authStore := authrepo.NewAuthRepository(dbConn)
	jwtMgr, err := securitytoken.NewJWT(securitytoken.Config{
		Secret:    appCfg.AuthJWTSecret,
		Issuer:    appCfg.AuthJWTIssuer,
		Audience:  appCfg.AuthJWTAudience,
		AccessTTL: appCfg.AccessTokenTTL,
	})
	if err != nil {
		return nil, fmt.Errorf("build jwt manager: %w", err)
	}

	authUseCase := authapp.NewAuthUseCase(authStore, jwtMgr, appCfg.RefreshTokenTTL, authapp.Dependencies{
		HashPassword:     crypto.HashPassword,
		ComparePassword:  crypto.ComparePassword,
		Now:              time.Now,
		RefreshTokenRand: rand.Reader,
	})

	mux := http.NewServeMux()
	userhttp.NewUserHandler(mux, userUseCase)
	authhttp.NewAuthHandler(mux, authUseCase, httpcookie.CookieConfig{
		Name:     appCfg.RefreshCookie,
		Path:     appCfg.RefreshPath,
		Secure:   appCfg.RefreshSecure,
		SameSite: httpcookie.ParseSameSite(appCfg.RefreshSameSite),
	})

	corsCfg := config.CORSConfig{
		AllowOrigin:  appCfg.CORSAllowOrigin,
		AllowMethods: appCfg.CORSAllowMethods,
		AllowHeaders: appCfg.CORSAllowHeaders,
	}
	httpHandler := middleware.CORSMiddleware(mux, corsCfg)
	httpHandler = middleware.RecoveryMiddleware(httpHandler)
	httpHandler = middleware.RequestLoggingMiddleware(httpHandler)
	httpHandler = middleware.RequestIDMiddleware(httpHandler)

	return httpHandler, nil
}

func NewServer(appCfg config.Config, httpHandler http.Handler) *http.Server {
	return &http.Server{
		Addr:    appCfg.ServerAddress,
		Handler: httpHandler,
	}
}
