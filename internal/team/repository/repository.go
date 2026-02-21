package repository

import (
	"context"
	"fmt"

	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/utils"
)

const ContextName = "Internal.Team.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	Create(context.Context, *model.Team) error
	Get(context.Context, string) *model.Team
	List(context.Context, schema.RequestTeamList) ([]model.Team, int64)
	Update(context.Context, *model.Team) error
	Delete(context.Context, string) error
}

func NewRepository(r Repository) IRepository {
	return &Repository{
		DB:    r.DB,
		Redis: r.Redis,
	}
}

func (r *Repository) Create(ctx context.Context, team *model.Team) error {
	return r.DB.GetTransaction(ctx).Create(team).Error
}

func (r *Repository) Get(ctx context.Context, id string) *model.Team {
	var team model.Team
	err := r.DB.GetTransaction(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&team).Error
	if err != nil {
		return nil
	}
	return &team
}

func (r *Repository) List(ctx context.Context, req schema.RequestTeamList) ([]model.Team, int64) {
	var teams []model.Team
	var total int64
	tx := r.DB.GetTransaction(ctx).Where("deleted_at IS NULL")

	tx.Model(&model.Team{}).Count(&total)

	offset := utils.CalculatePageSkip(req.Page, req.PageSize)
	sortBy := "name"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "asc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	tx = tx.Offset(offset).Limit(req.PageSize)
	tx = tx.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	tx.Find(&teams)

	return teams, total
}

func (r *Repository) Update(ctx context.Context, team *model.Team) error {
	return r.DB.GetTransaction(ctx).Save(team).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.GetTransaction(ctx).
		Model(&model.Team{}).
		Where("id = ?", id).
		Update("deleted_at", "NOW()").Error
}
