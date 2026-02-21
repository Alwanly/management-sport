package repository

import (
	"context"
	"fmt"

	"github.com/Alwanly/management-sport/internal/goal/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/utils"
)

const ContextName = "Internal.Goal.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	Create(context.Context, *model.Goal) error
	Get(context.Context, string) *model.Goal
	List(context.Context, schema.RequestGoalList) ([]model.Goal, int64)
	Delete(context.Context, string) error
}

func NewRepository(r Repository) IRepository {
	return &Repository{
		DB:    r.DB,
		Redis: r.Redis,
	}
}

func (r *Repository) Create(ctx context.Context, g *model.Goal) error {
	return r.DB.GetTransaction(ctx).Create(g).Error
}

func (r *Repository) Get(ctx context.Context, id string) *model.Goal {
	var g model.Goal
	err := r.DB.GetTransaction(ctx).Where("id = ?", id).First(&g).Error
	if err != nil {
		return nil
	}
	return &g
}

func (r *Repository) List(ctx context.Context, req schema.RequestGoalList) ([]model.Goal, int64) {
	var goals []model.Goal
	var total int64
	tx := r.DB.GetTransaction(ctx)

	tx.Model(&model.Goal{}).Count(&total)

	offset := utils.CalculatePageSkip(req.Page, req.PageSize)
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "asc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	tx = tx.Offset(offset).Limit(req.PageSize)
	tx = tx.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	tx.Find(&goals)

	return goals, total
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.GetTransaction(ctx).Where("id = ?", id).Delete(&model.Goal{}).Error
}
