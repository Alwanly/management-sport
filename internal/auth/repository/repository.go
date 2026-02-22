package repository

import (
	"context"

	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
)

const ContextName = "Internal.Auth.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	GetByUsername(context.Context, string) *model.User
	Create(context.Context, *model.User) error
	ExistsByUsername(context.Context, string) bool
}

func NewRepository(r Repository) IRepository {
	return &Repository{
		DB:    r.DB,
		Redis: r.Redis,
	}
}

func (r *Repository) GetByUsername(ctx context.Context, username string) *model.User {
	var user model.User
	err := r.DB.GetTransaction(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		return nil
	}
	return &user
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	return r.DB.GetTransaction(ctx).Create(user).Error
}

func (r *Repository) ExistsByUsername(ctx context.Context, username string) bool {
	var count int64
	r.DB.GetTransaction(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count)
	return count > 0
}
