package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/goal/repository"
	"github.com/Alwanly/management-sport/internal/goal/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Goal.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestGoalCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestGoalGet) wrapper.JSONResult
	List(context.Context, *schema.RequestGoalList) wrapper.JSONResult
	Delete(context.Context, *schema.RequestGoalDelete) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestGoalCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	now := time.Now()
	g := &model.Goal{
		ID:           utils.GenerateUUID(),
		MatchID:      req.MatchID,
		PlayerID:     req.PlayerID,
		MinuteScored: req.MinuteScored,
		CreatedAt:    now,
	}

	if err := u.Repository.Create(ctx, g); err != nil {
		l.Error("failed to create a goal", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to create a goal", nil)
	}

	l.Debug("goal created", zap.String("id", g.ID))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponseGoalCreate{ID: g.ID})
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestGoalGet) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Get"))

	g := u.Repository.Get(ctx, req.ID)
	if g == nil {
		l.Error("goal not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("GOAL_NOT_FOUND"), "Goal not found", nil)
	}

	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseGoalGet{
		ID:           g.ID,
		MatchID:      g.MatchID,
		PlayerID:     g.PlayerID,
		MinuteScored: g.MinuteScored,
	})
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestGoalList) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "List"))

	goals, total := u.Repository.List(ctx, *req)

	items := make([]schema.ResponseGoalItem, len(goals))
	for i, gg := range goals {
		items[i] = schema.ResponseGoalItem{
			ID:           gg.ID,
			MatchID:      gg.MatchID,
			PlayerID:     gg.PlayerID,
			MinuteScored: gg.MinuteScored,
		}
	}

	l.Debug("goals listed", zap.Int64("total", total))
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(goals), int(total), items, nil)
}

func (u *UseCase) Delete(ctx context.Context, req *schema.RequestGoalDelete) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Delete"))

	g := u.Repository.Get(ctx, req.ID)
	if g == nil {
		l.Error("goal not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("GOAL_NOT_FOUND"), "Goal not found", nil)
	}

	if err := u.Repository.Delete(ctx, req.ID); err != nil {
		l.Error("failed to delete a goal", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to delete a goal", nil)
	}

	l.Debug("goal deleted", zap.String("id", req.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseGoalDelete{})
}
