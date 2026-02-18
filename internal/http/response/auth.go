package response

import (
	"time"

	authusecase "admin.com/admin-api/internal/usecase/auth"
)

type SessionOutput struct {
	AccessToken      string     `json:"accessToken"`
	TokenType        string     `json:"tokenType"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	RefreshToken     string     `json:"-"`
	RefreshExpiresAt time.Time  `json:"-"`
	User             UserOutput `json:"user"`
}

func FromAuthSession(session authusecase.SessionOutput) SessionOutput {
	return SessionOutput{
		AccessToken: session.AccessToken,
		TokenType:   session.TokenType,
		ExpiresAt:   session.ExpiresAt,
		User:        FromAuthUser(session.User),
	}
}

func FromAuthUser(user authusecase.UserOutput) UserOutput {
	return UserOutput{
		ID:        user.ID,
		Name:      user.Name,
		LastName:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
