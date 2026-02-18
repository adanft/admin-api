package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"time"

	"admin.com/admin-api/internal/domain"
	domainauth "admin.com/admin-api/internal/domain/auth"
	userdomain "admin.com/admin-api/internal/domain/user"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*UserOutput, error)
	Login(ctx context.Context, input LoginInput) (*SessionOutput, error)
	Refresh(ctx context.Context, refreshToken string) (*SessionOutput, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, accessToken string) (*UserOutput, error)
}

type authUseCase struct {
	authRepo         domainauth.AuthRepository
	tokenManager     domainauth.AccessTokenManager
	refreshTokenTTL  time.Duration
	hashPassword     func(password string) (string, error)
	comparePassword  func(hash string, password string) error
	now              func() time.Time
	refreshTokenRand io.Reader
}

type Dependencies struct {
	HashPassword     func(password string) (string, error)
	ComparePassword  func(hash string, password string) error
	Now              func() time.Time
	RefreshTokenRand io.Reader
}

func NewAuthUseCase(
	authRepo domainauth.AuthRepository,
	tokenManager domainauth.AccessTokenManager,
	refreshTokenTTL time.Duration,
	dependencies Dependencies,
) AuthUseCase {
	if refreshTokenTTL <= 0 {
		refreshTokenTTL = 7 * 24 * time.Hour
	}
	if dependencies.HashPassword == nil {
		dependencies.HashPassword = func(string) (string, error) {
			return "", domain.ErrInternalServerError
		}
	}
	if dependencies.ComparePassword == nil {
		dependencies.ComparePassword = func(string, string) error {
			return domain.ErrInternalServerError
		}
	}
	if dependencies.Now == nil {
		dependencies.Now = time.Now
	}
	if dependencies.RefreshTokenRand == nil {
		dependencies.RefreshTokenRand = rand.Reader
	}

	return &authUseCase{
		authRepo:         authRepo,
		tokenManager:     tokenManager,
		refreshTokenTTL:  refreshTokenTTL,
		hashPassword:     dependencies.HashPassword,
		comparePassword:  dependencies.ComparePassword,
		now:              dependencies.Now,
		refreshTokenRand: dependencies.RefreshTokenRand,
	}
}

func (s *authUseCase) Register(ctx context.Context, input RegisterInput) (*UserOutput, error) {
	normalized, err := domainauth.NormalizeRegister(domainauth.RegisterData{
		Name:     input.Name,
		LastName: input.LastName,
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
		Avatar:   input.Avatar,
	})
	if err != nil {
		return nil, err
	}

	hashedPassword, err := s.hashPassword(normalized.Password)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	user, err := userdomain.NewUser(userdomain.UserProfile{
		Name:     normalized.Name,
		LastName: normalized.LastName,
		Username: normalized.Username,
		Email:    normalized.Email,
		Avatar:   normalized.Avatar,
	})
	if err != nil {
		return nil, err
	}
	user.PasswordHash = hashedPassword

	if err := s.authRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	userOut := toUserOutput(user)
	return &userOut, nil
}

func (s *authUseCase) Login(ctx context.Context, input LoginInput) (*SessionOutput, error) {
	normalized, err := domainauth.NormalizeLogin(domainauth.LoginData{
		Identity: input.Identity,
		Password: input.Password,
	})
	if err != nil {
		return nil, err
	}

	user, err := s.authRepo.GetUserByIdentity(ctx, normalized.Identity)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := s.comparePassword(user.PasswordHash, normalized.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.createSessionForUser(ctx, user, uuid.Nil)
}

func (s *authUseCase) Refresh(ctx context.Context, refreshToken string) (*SessionOutput, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, domain.ErrUnauthorized
	}

	now := s.now().UTC()
	refreshTokenHash := domainauth.HashRefreshToken(refreshToken)
	storedToken, err := s.authRepo.GetRefreshTokenByHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	if !storedToken.IsActiveAt(now) {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.authRepo.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	newRefreshToken, newRefreshTokenHash, err := s.generateRefreshTokenPair()
	if err != nil {
		return nil, err
	}

	refreshExpiresAt := now.Add(s.refreshTokenTTL)
	nextToken := domainauth.NewRefreshToken(storedToken.UserID, storedToken.FamilyID, newRefreshTokenHash, refreshExpiresAt)

	if err := s.authRepo.RotateRefreshToken(ctx, storedToken.ID, nextToken, now); err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrConflict) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	return &SessionOutput{
		AccessToken:      accessToken,
		TokenType:        "Bearer",
		ExpiresAt:        accessExpiresAt,
		RefreshToken:     newRefreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		User:             toUserOutput(user),
	}, nil
}

func (s *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}

	revokedAt := s.now().UTC()
	refreshTokenHash := domainauth.HashRefreshToken(refreshToken)
	if err := s.authRepo.RevokeRefreshTokenByHash(ctx, refreshTokenHash, revokedAt); err != nil {
		return err
	}

	return nil
}

func (s *authUseCase) Me(ctx context.Context, accessToken string) (*UserOutput, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, domain.ErrUnauthorized
	}

	claims, err := s.tokenManager.ParseAccessToken(accessToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	userOut := toUserOutput(user)
	return &userOut, nil
}

func (s *authUseCase) createSessionForUser(ctx context.Context, user *userdomain.User, familyID uuid.UUID) (*SessionOutput, error) {
	if user == nil || user.ID == uuid.Nil {
		return nil, domain.ErrInternalServerError
	}

	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	rawRefreshToken, refreshTokenHash, err := s.generateRefreshTokenPair()
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	if familyID == uuid.Nil {
		familyID = uuid.New()
	}

	refreshExpiresAt := now.Add(s.refreshTokenTTL)
	refreshToken := domainauth.NewRefreshToken(user.ID, familyID, refreshTokenHash, refreshExpiresAt)
	if err := s.authRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &SessionOutput{
		AccessToken:      accessToken,
		TokenType:        "Bearer",
		ExpiresAt:        accessExpiresAt,
		RefreshToken:     rawRefreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		User:             toUserOutput(user),
	}, nil
}

func (s *authUseCase) generateRefreshTokenPair() (string, string, error) {
	raw := make([]byte, 48)
	if _, err := io.ReadFull(s.refreshTokenRand, raw); err != nil {
		return "", "", domain.ErrInternalServerError
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(raw)
	return refreshToken, domainauth.HashRefreshToken(refreshToken), nil
}

func toUserOutput(user *userdomain.User) UserOutput {
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
