package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/match/repository"
	"github.com/Alwanly/management-sport/internal/match/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Match.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestMatchCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestMatchGet) wrapper.JSONResult
	List(context.Context, *schema.RequestMatchList) wrapper.JSONResult
	Update(context.Context, *schema.RequestMatchUpdate) wrapper.JSONResult
	Delete(context.Context, *schema.RequestMatchDelete) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestMatchCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	now := time.Now()
	m := &model.Match{
		ID:         utils.GenerateUUID(),
		MatchDate:  req.MatchDate,
		MatchTime:  req.MatchTime,
		HomeTeamID: req.HomeTeamID,
		AwayTeamID: req.AwayTeamID,
		HomeScore:  0,
		AwayScore:  0,
		Status:     model.MatchStatusScheduled,
		CreatedBy:  req.AuthUserData.UserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := u.Repository.Create(ctx, m); err != nil {
		l.Error("failed to create a match", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to create a match", nil)
	}

	l.Debug("match created", zap.String("id", m.ID))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponseMatchCreate{ID: m.ID})
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestMatchGet) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Get"))

	m := u.Repository.Get(ctx, req.ID)
	if m == nil {
		l.Error("match not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("MATCH_NOT_FOUND"), "Match not found", nil)
	}

	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseMatchGet{
		ID:         m.ID,
		MatchDate:  m.MatchDate,
		MatchTime:  m.MatchTime,
		HomeTeamID: m.HomeTeamID,
		AwayTeamID: m.AwayTeamID,
		HomeScore:  m.HomeScore,
		AwayScore:  m.AwayScore,
		Status:     string(m.Status),
	})
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestMatchList) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "List"))

	matches, total := u.Repository.List(ctx, *req)

	// convert to response items
	items := make([]schema.ResponseMatchItem, len(matches))
	for i, mm := range matches {
		items[i] = schema.ResponseMatchItem{
			ID:         mm.ID,
			MatchDate:  mm.MatchDate,
			HomeTeamID: mm.HomeTeamID,
			AwayTeamID: mm.AwayTeamID,
			Status:     string(mm.Status),
		}
	}

	l.Debug("matches listed", zap.Int64("total", total))
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(matches), int(total), items, nil)
}

func (u *UseCase) Update(ctx context.Context, req *schema.RequestMatchUpdate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Update"))

	m := u.Repository.Get(ctx, req.ID)
	if m == nil {
		l.Error("match not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("MATCH_NOT_FOUND"), "Match not found", nil)
	}

	m.MatchDate = req.MatchDate
	m.MatchTime = req.MatchTime
	m.HomeTeamID = req.HomeTeamID
	m.AwayTeamID = req.AwayTeamID
	if req.Status != "" {
		m.Status = model.MatchStatus(req.Status)
	}
	m.HomeScore = req.HomeScore
	m.AwayScore = req.AwayScore
	m.UpdatedAt = time.Now()
	m.UpdatedBy = req.AuthUserData.UserID

	if err := u.Repository.Update(ctx, m); err != nil {
		l.Error("failed to update a match", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to update a match", nil)
	}

	l.Debug("match updated", zap.String("id", m.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseMatchUpdate{ID: m.ID})
}

func (u *UseCase) Delete(ctx context.Context, req *schema.RequestMatchDelete) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Delete"))

	m := u.Repository.Get(ctx, req.ID)
	if m == nil {
		l.Error("match not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("MATCH_NOT_FOUND"), "Match not found", nil)
	}

	if err := u.Repository.Delete(ctx, req.ID); err != nil {
		l.Error("failed to delete a match", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to delete a match", nil)
	}

	l.Debug("match deleted", zap.String("id", req.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseMatchDelete{})
}
