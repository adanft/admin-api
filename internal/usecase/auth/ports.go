package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Name     string
	LastName string
	Username string
	Email    string
	Password string
	Avatar   string
}

type LoginInput struct {
	Identity string
	Password string
}

type UserOutput struct {
	ID        uuid.UUID
	Name      string
	LastName  string
	Username  string
	Email     string
	Avatar    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SessionOutput struct {
	AccessToken      string
	TokenType        string
	ExpiresAt        time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
	User             UserOutput
}
