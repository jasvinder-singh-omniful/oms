package service

import (
	"context"
	"time"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/storage/postgres"
	"github.com/si/internal/types"
	"github.com/si/internal/utils/hash"
)

type UserService struct {
	UserRepo *postgres.UserRepo
}

func NewUserService(userRepo *postgres.UserRepo) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, name string, email string, phone string, password_hash string) (*types.User, error) {
	logTag := "[UserService][CreateUser]"
	log.InfofWithContext(ctx, logTag+" creating user", "email", email)

	user := &types.User{
		Name:         name,
		Email:        email,
		Phone:        phone,
		PasswordHash: password_hash,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdUser, err := s.UserRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	log.InfofWithContext(ctx, logTag+" user created successfully", "email", createdUser.Email)
	return createdUser, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	logTag := "[UserService][GetUserByEmail]"
	log.InfofWithContext(ctx, logTag+" getting user by email", " email ", email)

	user, err := s.UserRepo.SearchByMail(ctx, email)

	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when getting user bt email", email)
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserById(ctx context.Context, id int64) (*types.User, error) {
	logTag := "[UserService][GetUserById]"
	log.InfofWithContext(ctx, logTag+" getting user by email", " email ", id)

	user, err := s.UserRepo.SearchByID(ctx, id)

	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when getting user bt id", id)
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, name string, email string, phone string, password string) (*types.User, error) {
	logTag := "[UserService][UpdateUser]"
	log.InfofWithContext(ctx, logTag+" updating user", "id", id)

	var hashedPassword string
	var err error
	if password != "" {
		hashedPassword, err = hash.HashPassword(password)
		if err != nil {
			log.ErrorfWithContext(ctx, logTag+" error when hashing password", err)
			return nil, err
		}
	}

	existingUser, err := s.UserRepo.SearchByID(ctx, id)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to fetch existing user", err)
		return nil, err
	}
	if existingUser == nil {
		log.WarnfWithContext(ctx, logTag+" user not found", "id", id)
		return nil, err
	}

	if name != "" {
		existingUser.Name = name
	}
	if email != "" {
		existingUser.Email = email
	}
	if phone != "" {
		existingUser.Phone = phone
	}
	if hashedPassword != "" {
		existingUser.PasswordHash = hashedPassword
	}

	updatedUser, err := s.UserRepo.Update(ctx, existingUser)
	if err != nil {
		return nil, err
	}
	log.InfofWithContext(ctx, logTag+" user updated successfully", "id", updatedUser.ID)
	return updatedUser, nil
}


func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	logTag := "[UserService][DeleteUser]"
	log.InfofWithContext(ctx, logTag+" deleting user", "user_id", id)


	err := s.UserRepo.Delete(ctx, id)

	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when deleting user", err, "user_id", id)
		return err
	}

	log.InfofWithContext(ctx, logTag+" user deleted successfully", "user_id", id)
	return nil
}