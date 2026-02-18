package auth

import (
	"net/http"
	"strings"

	"admin.com/admin-api/internal/domain"
	httpcookie "admin.com/admin-api/internal/http/cookie"
	"admin.com/admin-api/internal/http/decoder"
	httprequest "admin.com/admin-api/internal/http/request"
	"admin.com/admin-api/internal/http/response"
	authusecase "admin.com/admin-api/internal/usecase/auth"
)

const authorizationBearerPrefix = "Bearer "

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req httprequest.RegisterInput
	if err := decoder.DecodeBody(w, r, &req); err != nil {
		decoder.WriteDecodeError(w, err)
		return
	}

	user, err := h.useCase.Register(r.Context(), authusecase.RegisterInput{
		Name:     req.Name,
		LastName: req.LastName,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Avatar:   req.Avatar,
	})
	if err != nil {
		writeAuthBusinessError(w, r, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, response.FromAuthUser(*user))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req httprequest.LoginInput
	if err := decoder.DecodeBody(w, r, &req); err != nil {
		decoder.WriteDecodeError(w, err)
		return
	}

	session, err := h.useCase.Login(r.Context(), authusecase.LoginInput{
		Identity: req.Identity,
		Password: req.Password,
	})
	if err != nil {
		writeAuthBusinessError(w, r, err)
		return
	}

	h.writeSession(w, http.StatusOK, session)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, ok := h.refreshTokenFromCookie(r)
	if !ok {
		writeAuthBusinessError(w, r, domain.ErrUnauthorized)
		return
	}

	session, err := h.useCase.Refresh(r.Context(), refreshToken)
	if err != nil {
		writeAuthBusinessError(w, r, err)
		return
	}

	h.writeSession(w, http.StatusOK, session)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if refreshToken, ok := h.refreshTokenFromCookie(r); ok {
		if err := h.useCase.Logout(r.Context(), refreshToken); err != nil {
			writeAuthBusinessError(w, r, err)
			return
		}
	}

	httpcookie.ClearRefreshToken(w, h.cookieConfig)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := accessTokenFromAuthorization(r.Header.Get("Authorization"))
	if !ok {
		writeAuthBusinessError(w, r, domain.ErrUnauthorized)
		return
	}

	user, err := h.useCase.Me(r.Context(), accessToken)
	if err != nil {
		writeAuthBusinessError(w, r, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, response.FromAuthUser(*user))
}

func (h *AuthHandler) writeSession(w http.ResponseWriter, status int, session *authusecase.SessionOutput) {
	httpcookie.SetRefreshToken(w, h.cookieConfig, session.RefreshToken, session.RefreshExpiresAt)
	response.WriteSuccess(w, status, response.FromAuthSession(*session))
}

func (h *AuthHandler) refreshTokenFromCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(h.cookieConfig.Name)
	if err != nil {
		return "", false
	}

	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return "", false
	}

	return token, true
}

func accessTokenFromAuthorization(authorizationHeader string) (string, bool) {
	authorizationHeader = strings.TrimSpace(authorizationHeader)
	if authorizationHeader == "" {
		return "", false
	}

	if !strings.HasPrefix(authorizationHeader, authorizationBearerPrefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, authorizationBearerPrefix))
	if token == "" {
		return "", false
	}

	return token, true
}
