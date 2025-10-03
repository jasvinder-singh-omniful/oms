package service

import (
	"context"

	"github.com/si/internal/storage/postgres"
	"github.com/si/internal/types"
)

type UserService struct {
	UserRepo *postgres.UserRepo
}

func NewUserService(userRepo *postgres.UserRepo) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *types.User) error {
	return s.UserRepo.CreateUser(ctx, user)
}

