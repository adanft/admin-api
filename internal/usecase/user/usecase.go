package user

import (
	"context"

	"admin.com/admin-api/internal/domain"
	userdomain "admin.com/admin-api/internal/domain/user"
	"github.com/google/uuid"
)

type UserUseCase interface {
	GetUser(ctx context.Context, id uuid.UUID) (*UserOutput, error)
	CreateUser(ctx context.Context, input CreateUserInput) (*UserOutput, error)
	GetUsers(ctx context.Context) ([]UserOutput, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdateUser(ctx context.Context, input UpdateUserInput) (*UserOutput, error)
}

type userUseCase struct {
	userRepo     userdomain.UserRepository
	hashPassword func(password string) (string, error)
}

func NewUserUseCase(userRepo userdomain.UserRepository, hashPassword func(password string) (string, error)) UserUseCase {
	if hashPassword == nil {
		hashPassword = func(string) (string, error) {
			return "", domain.ErrInternalServerError
		}
	}

	return &userUseCase{
		userRepo:     userRepo,
		hashPassword: hashPassword,
	}
}

func (s *userUseCase) GetUser(ctx context.Context, id uuid.UUID) (*UserOutput, error) {
	if id == uuid.Nil {
		return nil, domain.ErrBadRequest
	}

	user, err := s.userRepo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	userOut := toUserOutput(user)
	return &userOut, nil
}

func (s *userUseCase) CreateUser(ctx context.Context, input CreateUserInput) (*UserOutput, error) {
	user, err := userdomain.NewUser(userdomain.UserProfile{
		Name:     input.Name,
		LastName: input.LastName,
		Username: input.Username,
		Email:    input.Email,
		Avatar:   input.Avatar,
	})
	if err != nil {
		return nil, err
	}

	if err := userdomain.SetTemporaryPassword(user, s.hashPassword); err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	userOut := toUserOutput(user)
	return &userOut, nil
}

func (s *userUseCase) GetUsers(ctx context.Context) ([]UserOutput, error) {
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	userOutputs := make([]UserOutput, len(users))
	for i := range users {
		userOutputs[i] = toUserOutput(&users[i])
	}

	return userOutputs, nil
}

func (s *userUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrBadRequest
	}

	return s.userRepo.DeleteUser(ctx, id)
}

func (s *userUseCase) UpdateUser(ctx context.Context, input UpdateUserInput) (*UserOutput, error) {
	if input.ID == uuid.Nil {
		return nil, domain.ErrBadRequest
	}

	user := &userdomain.User{ID: input.ID}
	if err := user.SetProfile(userdomain.UserProfile{
		Name:     input.Name,
		LastName: input.LastName,
		Username: input.Username,
		Email:    input.Email,
		Avatar:   input.Avatar,
	}); err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	updatedUser, err := s.userRepo.GetUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	userOut := toUserOutput(updatedUser)
	return &userOut, nil
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
