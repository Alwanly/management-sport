package repository

import (
	"context"
	"fmt"

	"github.com/Alwanly/management-sport/internal/match/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/utils"
)

const ContextName = "Internal.Match.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	Create(context.Context, *model.Match) error
	Get(context.Context, string) *model.Match
	List(context.Context, schema.RequestMatchList) ([]model.Match, int64)
	Update(context.Context, *model.Match) error
	Delete(context.Context, string) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

func NewRepository(r Repository) IRepository {
	return &Repository{
		DB:    r.DB,
		Redis: r.Redis,
	}
}

func (r *Repository) Create(ctx context.Context, m *model.Match) error {
	return r.DB.GetTransaction(ctx).Create(m).Error
}

func (r *Repository) Get(ctx context.Context, id string) *model.Match {
	var m model.Match
	err := r.DB.GetTransaction(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil
	}
	return &m
}

func (r *Repository) List(ctx context.Context, req schema.RequestMatchList) ([]model.Match, int64) {
	var matches []model.Match
	var total int64
	tx := r.DB.GetTransaction(ctx)

	tx.Model(&model.Match{}).Count(&total)

	offset := utils.CalculatePageSkip(req.Page, req.PageSize)
	sortBy := "match_date"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "asc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	tx = tx.Offset(offset).Limit(req.PageSize)
	tx = tx.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	tx.Find(&matches)

	return matches, total
}

func (r *Repository) Update(ctx context.Context, m *model.Match) error {
	return r.DB.GetTransaction(ctx).Save(m).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.GetTransaction(ctx).Where("id = ?", id).Delete(&model.Match{}).Error
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.DB.GetTransaction(ctx).Model(&model.Match{}).Where("id = ?", id).Update("status", status).Error
}
