package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/player/repository"
	"github.com/Alwanly/management-sport/internal/player/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Player.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestPlayerCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestPlayerGet) wrapper.JSONResult
	List(context.Context, *schema.RequestPlayerList) wrapper.JSONResult
	Update(context.Context, *schema.RequestPlayerUpdate) wrapper.JSONResult
	Delete(context.Context, *schema.RequestPlayerDelete) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestPlayerCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	now := time.Now()
	player := &model.Player{
		ID:          utils.GenerateUUID(),
		TeamID:      req.TeamID,
		Name:        req.Name,
		HeightCM:    req.HeightCM,
		WeightKG:    req.WeightKG,
		Position:    model.PlayerPosition(req.Position),
		ShirtNumber: req.ShirtNumber,
		CreatedBy:   req.AuthUserData.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.Repository.Create(ctx, player); err != nil {
		l.Error("failed to create a player", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to create a player", nil)
	}

	l.Debug("player created", zap.String("id", player.ID))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponsePlayerCreate{ID: player.ID})
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestPlayerGet) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Get"))

	p := u.Repository.Get(ctx, req.ID)
	if p == nil {
		l.Error("player not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("PLAYER_NOT_FOUND"), "Player not found", nil)
	}

	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponsePlayerGet{
		ID:          p.ID,
		TeamID:      p.TeamID,
		Name:        p.Name,
		HeightCM:    p.HeightCM,
		WeightKG:    p.WeightKG,
		Position:    string(p.Position),
		ShirtNumber: p.ShirtNumber,
	})
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestPlayerList) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "List"))

	players, total := u.Repository.List(ctx, *req)

	// convert to response items
	items := make([]schema.ResponsePlayerItem, len(players))
	for i, p := range players {
		items[i] = schema.ResponsePlayerItem{
			ID:          p.ID,
			TeamID:      p.TeamID,
			Name:        p.Name,
			Position:    string(p.Position),
			ShirtNumber: p.ShirtNumber,
		}
	}

	l.Debug("players listed", zap.Int64("total", total))
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(players), int(total), items, nil)
}

func (u *UseCase) Update(ctx context.Context, req *schema.RequestPlayerUpdate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Update"))

	p := u.Repository.Get(ctx, req.ID)
	if p == nil {
		l.Error("player not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("PLAYER_NOT_FOUND"), "Player not found", nil)
	}

	p.Name = req.Name
	p.HeightCM = req.HeightCM
	p.WeightKG = req.WeightKG
	p.Position = model.PlayerPosition(req.Position)
	p.ShirtNumber = req.ShirtNumber
	p.UpdatedAt = time.Now()
	p.UpdatedBy = req.AuthUserData.UserID

	if err := u.Repository.Update(ctx, p); err != nil {
		l.Error("failed to update a player", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to update a player", nil)
	}

	l.Debug("player updated", zap.String("id", p.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponsePlayerUpdate{ID: p.ID})
}

func (u *UseCase) Delete(ctx context.Context, req *schema.RequestPlayerDelete) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Delete"))

	p := u.Repository.Get(ctx, req.ID)
	if p == nil {
		l.Error("player not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("PLAYER_NOT_FOUND"), "Player not found", nil)
	}

	if err := u.Repository.Delete(ctx, req.ID); err != nil {
		l.Error("failed to delete a player", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to delete a player", nil)
	}

	l.Debug("player deleted", zap.String("id", req.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponsePlayerDelete{})
}
