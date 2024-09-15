package service

import (
	"context"
	"fmt"
	"log/slog"

	"tender-service/internal/entity"
	"tender-service/internal/repo"
	"tender-service/internal/repo/repoerrs"
)

type UserService struct {
	userRepo repo.User
}

func NewUserService(userRepo repo.User) *UserService {
	return &UserService{userRepo: userRepo}
}

func (u *UserService) Create(ctx context.Context, log *slog.Logger, input UserCreateInput) (string, error) {
	log.Info(fmt.Sprintf("Service - UserService - Create"))
	user := entity.User{
		Username:  input.Username,
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}
	id, err := u.userRepo.Create(ctx, user)
	if err != nil {
		if err == repoerrs.ErrAlreadyExists {
			return "", ErrUserAlreadyExists
		}
		log.Error(fmt.Sprintf("Service - UserService - Create: %v", err))
		return "", ErrCannotCreateUser
	}
	log.Info(fmt.Sprintf("Service - UserService - userRepo.Create - id: %s", id))
	return id, nil
}

func (u *UserService) GetById(ctx context.Context, log *slog.Logger, input UserGetByIdInput) (entity.User, error) {
	user, err := u.userRepo.GetById(ctx, input.Id)
	if err != nil {
		if err == repoerrs.ErrNotFound {
			return entity.User{}, ErrUserNotFound
		}
		log.Error(fmt.Sprintf("Service - UserService - GetById: %v", err))
		return entity.User{}, ErrCannotGetUser
	}
	return user, nil
}

func (u *UserService) GetByUsername(ctx context.Context, log *slog.Logger, input UserGetByUsernameInput) (
	entity.User, error,
) {
	user, err := u.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		if err == repoerrs.ErrNotFound {
			return entity.User{}, ErrUserNotFound
		}
		log.Error(fmt.Sprintf("Service - UserService - GetByUsername: %v", err))
		return entity.User{}, ErrCannotGetUser
	}
	return user, nil
}
