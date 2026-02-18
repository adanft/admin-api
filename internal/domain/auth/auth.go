package auth

import (
	"time"

	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	Subject   string
	Audience  string
	Issuer    string
	IssuedAt  time.Time
	ExpiresAt time.Time
	TokenID   string
}

type AccessTokenManager interface {
	GenerateAccessToken(userID uuid.UUID) (string, time.Time, error)
	ParseAccessToken(token string) (*AccessTokenClaims, error)
}

type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	FamilyID   uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

func NewRefreshToken(userID uuid.UUID, familyID uuid.UUID, tokenHash string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt.UTC(),
	}
}

func (t *RefreshToken) IsActiveAt(now time.Time) bool {
	if t == nil {
		return false
	}
	if t.RevokedAt != nil {
		return false
	}

	return t.ExpiresAt.After(now.UTC())
}
