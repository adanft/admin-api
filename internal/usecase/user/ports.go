package user

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	Name     string
	LastName string
	Username string
	Email    string
	Avatar   string
}

type UpdateUserInput struct {
	ID       uuid.UUID
	Name     string
	LastName string
	Username string
	Email    string
	Avatar   string
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
