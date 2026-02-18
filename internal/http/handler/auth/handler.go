package auth

import (
	"net/http"

	httpcookie "admin.com/admin-api/internal/http/cookie"
	authusecase "admin.com/admin-api/internal/usecase/auth"
)

type AuthHandler struct {
	useCase      authusecase.AuthUseCase
	cookieConfig httpcookie.CookieConfig
}

func NewAuthHandler(mux *http.ServeMux, useCase authusecase.AuthUseCase, cookieConfig httpcookie.CookieConfig) *AuthHandler {
	handler := &AuthHandler{
		useCase:      useCase,
		cookieConfig: cookieConfig,
	}

	handler.RegisterRoutes(mux)
	return handler
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
	mux.HandleFunc("POST /auth/logout", h.Logout)
	mux.HandleFunc("GET /auth/me", h.Me)
}
