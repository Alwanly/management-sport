package repository

import (
	"context"
	"fmt"

	"github.com/Alwanly/management-sport/internal/player/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/utils"
	"gorm.io/gorm"
)

const ContextName = "Internal.Player.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	Create(context.Context, *model.Player) error
	Get(context.Context, string) *model.Player
	List(context.Context, schema.RequestPlayerList) ([]model.Player, int64)
	Update(context.Context, *model.Player) error
	Delete(context.Context, string) error
	GetTeamName(ctx context.Context, teamID string) (string, error)
}

func NewRepository(r Repository) IRepository {
	return &Repository{
		DB:    r.DB,
		Redis: r.Redis,
	}
}

func (r *Repository) Create(ctx context.Context, player *model.Player) error {
	return r.DB.GetTransaction(ctx).Create(player).Error
}

func (r *Repository) Get(ctx context.Context, id string) *model.Player {
	var p model.Player
	err := r.DB.GetTransaction(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&p).Error
	if err != nil {
		return nil
	}
	return &p
}

func (r *Repository) List(ctx context.Context, req schema.RequestPlayerList) ([]model.Player, int64) {
	var players []model.Player
	var total int64
	tx := r.DB.GetTransaction(ctx).
		Model(&model.Player{}).
		Joins("Team").
		Where("players.deleted_at IS NULL")

	// Count using a separate session to avoid shared state
	tx.Session(&gorm.Session{}).Count(&total)

	offset := utils.CalculatePageSkip(req.Page, req.PageSize)

	sortBy := "name"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "asc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	tx.Session(&gorm.Session{}).
		Offset(offset).
		Limit(req.PageSize).
		Order(fmt.Sprintf("players.%s %s", sortBy, sortOrder)).
		Find(&players)

	return players, total
}

func (r *Repository) Update(ctx context.Context, player *model.Player) error {
	return r.DB.GetTransaction(ctx).Save(player).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.GetTransaction(ctx).
		Model(&model.Player{}).
		Where("id = ?", id).
		Update("deleted_at", "NOW()").Error
}

func (r *Repository) GetTeamName(ctx context.Context, teamID string) (string, error) {
	var teamName string
	err := r.DB.GetTransaction(ctx).
		Model(&model.Team{}).
		Select("name").
		Where("id = ? AND deleted_at IS NULL", teamID).
		First(&teamName).Error
	if err != nil {
		return "", err
	}
	return teamName, nil
}
