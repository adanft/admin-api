package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type DBUser struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name         string    `bun:"name,notnull"`
	LastName     string    `bun:"last_name,notnull"`
	Username     string    `bun:"username,unique,notnull"`
	PasswordHash string    `bun:"password_hash,notnull"`
	Email        string    `bun:"email,notnull"`
	Avatar       string    `bun:"avatar"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type DBRole struct {
	bun.BaseModel `bun:"table:roles,alias:r"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name        string    `bun:"name,unique,notnull"`
	Description string    `bun:"description"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

type DBUserRole struct {
	bun.BaseModel `bun:"table:user_roles,alias:ur"`

	UserID     uuid.UUID `bun:"user_id,pk,type:uuid,notnull"`
	RoleID     uuid.UUID `bun:"role_id,pk,type:uuid,notnull"`
	AssignedAt time.Time `bun:"assigned_at,nullzero,notnull,default:current_timestamp"`
}
