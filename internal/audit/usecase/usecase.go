package usecase

import (
	"context"
	"net/http"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/audit/repository"
	"github.com/Alwanly/management-sport/internal/audit/schema"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Audit.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Get(context.Context, *schema.RequestAuditGet) wrapper.JSONResult
	List(context.Context, *schema.RequestAuditList) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{Config: uc.Config, Logger: uc.Logger, Repository: uc.Repository}
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestAuditGet) wrapper.JSONResult {
	a := u.Repository.Get(ctx, req.ID)
	if a == nil {
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("AUDIT_NOT_FOUND"), "Audit not found", nil)
	}

	resp := schema.ResponseAuditGet{
		ID:         a.ID,
		EntityType: a.EntityType,
		EntityID:   a.EntityID,
		Action:     string(a.Action),
		OldValue:   a.OldValue,
		NewValue:   a.NewValue,
		UserID:     a.UserID,
		UserRole:   a.UserRole,
		CreatedAt:  a.CreatedAt,
	}
	return wrapper.ResponseSuccess(200, resp)
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestAuditList) wrapper.JSONResult {
	items, total := u.Repository.List(ctx, *req)
	resp := make([]schema.ResponseAuditItem, len(items))
	for i, a := range items {
		resp[i] = schema.ResponseAuditItem{
			ID:         a.ID,
			EntityType: a.EntityType,
			EntityID:   a.EntityID,
			Action:     string(a.Action),
			UserID:     a.UserID,
			UserRole:   a.UserRole,
			CreatedAt:  a.CreatedAt,
		}
	}
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(resp), int(total), resp, nil)
}
