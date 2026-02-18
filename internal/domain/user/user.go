package user

import (
	"strings"
	"time"

	"admin.com/admin-api/internal/domain"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	LastName     string
	Username     string
	Email        string
	Avatar       string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserProfile struct {
	Name     string
	LastName string
	Username string
	Email    string
	Avatar   string
}

func NewUser(profile UserProfile) (*User, error) {
	user := &User{}
	if err := user.SetProfile(profile); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) SetProfile(profile UserProfile) error {
	name := strings.TrimSpace(profile.Name)
	lastName := strings.TrimSpace(profile.LastName)
	username := strings.TrimSpace(profile.Username)
	avatar := strings.TrimSpace(profile.Avatar)

	if name == "" || lastName == "" || username == "" {
		return domain.ErrBadRequest
	}

	email, err := domain.NewEmail(profile.Email)
	if err != nil {
		return err
	}

	u.Name = name
	u.LastName = lastName
	u.Username = username
	u.Email = email.String()
	u.Avatar = avatar
	return nil
}
