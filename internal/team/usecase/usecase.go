package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/team/repository"
	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Team.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestTeamCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestTeamGet) wrapper.JSONResult
	List(context.Context, *schema.RequestTeamList) wrapper.JSONResult
	Update(context.Context, *schema.RequestTeamUpdate) wrapper.JSONResult
	Delete(context.Context, *schema.RequestTeamDelete) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestTeamCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	now := time.Now()
	team := &model.Team{
		ID:          utils.GenerateUUID(),
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		FoundedYear: req.FoundedYear,
		Address:     req.Address,
		City:        req.City,
		CreatedBy:   req.AuthUserData.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.Repository.Create(ctx, team); err != nil {
		l.Error("failed to create a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to create a team", nil)
	}

	l.Debug("team created", zap.String("id", team.ID))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponseTeamCreate{ID: team.ID})
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestTeamGet) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Get"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamGet{
		ID:          team.ID,
		Name:        team.Name,
		LogoURL:     team.LogoURL,
		FoundedYear: team.FoundedYear,
		Address:     team.Address,
		City:        team.City,
	})
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestTeamList) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "List"))

	teams, total := u.Repository.List(ctx, *req)

	response := req.ToResponse(teams)
	l.Debug("teams listed", zap.Int64("total", total))
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(teams), int(total), response, nil)
}

func (u *UseCase) Update(ctx context.Context, req *schema.RequestTeamUpdate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Update"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	team.Name = req.Name
	team.LogoURL = req.LogoURL
	team.FoundedYear = req.FoundedYear
	team.Address = req.Address
	team.City = req.City
	team.UpdatedAt = time.Now()
	team.UpdatedBy = req.AuthUserData.UserID

	if err := u.Repository.Update(ctx, team); err != nil {
		l.Error("failed to update a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to update a team", nil)
	}

	l.Debug("team updated", zap.String("id", team.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamUpdate{ID: team.ID})
}

func (u *UseCase) Delete(ctx context.Context, req *schema.RequestTeamDelete) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Delete"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	if err := u.Repository.Delete(ctx, req.ID); err != nil {
		l.Error("failed to delete a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to delete a team", nil)
	}

	l.Debug("team deleted", zap.String("id", req.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamDelete{})
}
