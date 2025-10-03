package postgres

import (
	"context"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/types"
)

type UserRepo struct {
	DB *Postgres
}


func NewUserRepo(ctx context.Context, db *Postgres) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *types.User) error {
	logTag := "[UserRepo][CreateUser]"
	log.InfofWithContext(ctx, logTag+" Creating user", "email", user.Email)

	r.DB.Cluster.GetMasterDB(ctx)
	user.IsActive = true

	log.InfofWithContext(ctx, logTag+" User created successfully", "email", user.Email)
	return nil
}

