package auth

import (
	"context"
	"time"

	userdomain "admin.com/admin-api/internal/domain/user"
	"github.com/google/uuid"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *userdomain.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error)
	GetUserByIdentity(ctx context.Context, identity string) (*userdomain.User, error)
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RotateRefreshToken(ctx context.Context, currentTokenID uuid.UUID, nextToken *RefreshToken, usedAt time.Time) error
	RevokeRefreshTokenByHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
}
