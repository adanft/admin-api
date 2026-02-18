package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type DBRefreshToken struct {
	bun.BaseModel `bun:"table:auth_refresh_tokens,alias:art"`

	ID         uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID     uuid.UUID  `bun:"user_id,type:uuid,notnull"`
	FamilyID   uuid.UUID  `bun:"family_id,type:uuid,notnull"`
	TokenHash  string     `bun:"token_hash,notnull"`
	ExpiresAt  time.Time  `bun:"expires_at,notnull"`
	RevokedAt  *time.Time `bun:"revoked_at"`
	LastUsedAt *time.Time `bun:"last_used_at"`
	CreatedAt  time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
