package repository

import (
	"context"
	"fmt"

	"github.com/Alwanly/management-sport/internal/audit/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/utils"
)

const ContextName = "Internal.Audit.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	Create(context.Context, *model.AuditLog) error
	Get(context.Context, string) *model.AuditLog
	List(context.Context, schema.RequestAuditList) ([]model.AuditLog, int64)
}

func NewRepository(r Repository) IRepository {
	return &Repository{DB: r.DB, Redis: r.Redis}
}

func (r *Repository) Create(ctx context.Context, a *model.AuditLog) error {
	return r.DB.GetTransaction(ctx).Create(a).Error
}

func (r *Repository) Get(ctx context.Context, id string) *model.AuditLog {
	var a model.AuditLog
	err := r.DB.GetTransaction(ctx).Where("id = ?", id).First(&a).Error
	if err != nil {
		return nil
	}
	return &a
}

func (r *Repository) List(ctx context.Context, req schema.RequestAuditList) ([]model.AuditLog, int64) {
	var items []model.AuditLog
	var total int64
	tx := r.DB.GetTransaction(ctx)

	if !req.From.IsZero() {
		tx = tx.Where("created_at >= ?", req.From)
	}
	if !req.To.IsZero() {
		tx = tx.Where("created_at <= ?", req.To)
	}
	if req.Entity != "" {
		tx = tx.Where("entity_type = ?", req.Entity)
	}
	if req.Action != "" {
		tx = tx.Where("action = ?", req.Action)
	}

	tx.Model(&model.AuditLog{}).Count(&total)

	offset := utils.CalculatePageSkip(req.Page, req.PageSize)
	tx = tx.Offset(offset).Limit(req.PageSize)
	tx = tx.Order(fmt.Sprintf("%s %s", "created_at", "desc"))
	tx.Find(&items)

	return items, total
}
