package auth

import (
	"crypto/sha256"
	"encoding/hex"

	"admin.com/admin-api/internal/domain"
	userdomain "admin.com/admin-api/internal/domain/user"
)

type RegisterData struct {
	Name     string
	LastName string
	Username string
	Email    string
	Password string
	Avatar   string
}

type LoginData struct {
	Identity string
	Password string
}

func NormalizeRegister(data RegisterData) (RegisterData, error) {
	user, err := userdomain.NewUser(userdomain.UserProfile{
		Name:     data.Name,
		LastName: data.LastName,
		Username: data.Username,
		Email:    data.Email,
		Avatar:   data.Avatar,
	})
	if err != nil {
		return RegisterData{}, err
	}

	password := domain.TrimPassword(data.Password)
	if password == "" {
		return RegisterData{}, domain.ErrBadRequest
	}

	if err := domain.ValidatePassword(password); err != nil {
		return RegisterData{}, err
	}

	return RegisterData{
		Name:     user.Name,
		LastName: user.LastName,
		Username: user.Username,
		Email:    user.Email,
		Password: password,
		Avatar:   user.Avatar,
	}, nil
}

func NormalizeLogin(data LoginData) (LoginData, error) {
	identity := domain.NormalizeIdentity(data.Identity)
	password := domain.TrimPassword(data.Password)
	if identity == "" || password == "" {
		return LoginData{}, domain.ErrBadRequest
	}

	return LoginData{
		Identity: identity,
		Password: password,
	}, nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
