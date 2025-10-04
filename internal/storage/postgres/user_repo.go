package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/types"
	"gorm.io/gorm"
)

type UserRepo struct {
	DB *Postgres
}


func NewUserRepo(ctx context.Context, db *Postgres) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *types.User) (*types.User, error) {
	logTag := "[UserRepo][CreateUser]"
	log.InfofWithContext(ctx, logTag+" Creating user", "email", user.Email)

	db := r.DB.Cluster.GetMasterDB(ctx)

	if err := db.Create(user).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to create user", err, "email", user.Email)
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" user created successfully", "email", user.Email)
	return user, nil
}


func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	logTag := "[UserRepo][GetUserByEmail]"
	log.InfofWithContext(ctx, logTag+" getting user by email", "email ", email)

	db := r.DB.Cluster.GetSlaveDB(ctx)

	var user types.User
	err := db.Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
            log.WarnfWithContext(ctx, logTag+" user not found", "email", email)
            return nil, nil
        }
		log.ErrorfWithContext(ctx, logTag+" error when getting user by email", err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetUserById(ctx context.Context, id int64) (*types.User, error) {
	logTag := "[UserRepo][GetUserById]"
	log.InfofWithContext(ctx, logTag+" getting user by id", "id", id)

	db := r.DB.Cluster.GetSlaveDB(ctx)

	var user types.User
	err := db.Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
            log.WarnfWithContext(ctx, logTag+" User not found", "id", id)
            return nil, nil
        }
		log.ErrorfWithContext(ctx, logTag+" error when getting user by id", err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, user *types.User) (*types.User, error) {
	logTag := "[UserRepo][UpdateUser]"
	log.InfofWithContext(ctx, logTag+" updating user", "id", user.ID)

	db := r.DB.Cluster.GetMasterDB(ctx)

	if err := db.Save(user).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to update user", err, "id", user.ID)
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" user updated successfully", "id", user.ID)
	return user, nil
}

func (r *UserRepo) DeleteUserById(ctx context.Context, id int64) error {
	logTag := "[UserRepo][DeleteUser]"
	log.InfofWithContext(ctx, logTag+" deleting user", "id", id)

	db := r.DB.Cluster.GetMasterDB(ctx)

	res := db.Model(&types.User{}).
		Where("id = ? AND is_active = ?", id, true).
		Updates(map[string]interface{}{
			"is_active": false,
			"updated_at": time.Now(),
		})

	if err := res.Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when dleting user")
		return fmt.Errorf("error when deleting user %v", err)
	}

	if res.RowsAffected == 0 {
		log.WarnfWithContext(ctx, logTag+" user not found or already deleted", "user_id", id)
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}
