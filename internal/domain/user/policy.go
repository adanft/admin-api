package user

import (
	"strings"

	"admin.com/admin-api/internal/domain"
)

func SetTemporaryPassword(user *User, hashPassword func(password string) (string, error)) error {
	if user == nil || strings.TrimSpace(user.Username) == "" {
		return domain.ErrBadRequest
	}
	if hashPassword == nil {
		return domain.ErrInternalServerError
	}

	hashedPassword, err := hashPassword(user.Username)
	if err != nil {
		return domain.ErrInternalServerError
	}

	user.PasswordHash = hashedPassword
	return nil
}
