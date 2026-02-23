package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/goal/repository"
	"github.com/Alwanly/management-sport/internal/goal/schema"
	matchRepository "github.com/Alwanly/management-sport/internal/match/repository"
	playerRepository "github.com/Alwanly/management-sport/internal/player/repository"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Goal.Usecase"

type UseCase struct {
	Config           *config.GlobalConfig
	Logger           *zap.Logger
	Repository       repository.IRepository
	MatchRepository  matchRepository.IRepository
	PlayerRepository playerRepository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestGoalCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestGoalGet) wrapper.JSONResult
	List(context.Context, *schema.RequestGoalList) wrapper.JSONResult
	Delete(context.Context, *schema.RequestGoalDelete) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:           uc.Config,
		Logger:           uc.Logger,
		Repository:       uc.Repository,
		MatchRepository:  uc.MatchRepository,
		PlayerRepository: uc.PlayerRepository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestGoalCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	// Step 1: Validate match exists and status is ongoing
	match := u.MatchRepository.Get(ctx, req.MatchID)
	if match == nil {
		l.Error("match not found", zap.String("match_id", req.MatchID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("MATCH_NOT_FOUND"), "Match not found", nil)
	}

	if match.Status != model.MatchStatusOngoing {
		l.Error("match is not ongoing", zap.String("match_id", req.MatchID), zap.String("status", string(match.Status)))
		return wrapper.ResponseFailed(http.StatusBadRequest, contract.CreateStatusCode("MATCH_NOT_ONGOING"), "Match is not ongoing", nil)
	}

	// Step 2: Validate player exists and belongs to one of the teams in the match
	player := u.PlayerRepository.Get(ctx, req.PlayerID)
	if player == nil {
		l.Error("player not found", zap.String("player_id", req.PlayerID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("PLAYER_NOT_FOUND"), "Player not found", nil)
	}

	if player.TeamID != match.HomeTeamID && player.TeamID != match.AwayTeamID {
		l.Error("player not in match teams", zap.String("player_id", player.ID), zap.String("player_team_id", player.TeamID), zap.String("home_team_id", match.HomeTeamID), zap.String("away_team_id", match.AwayTeamID))
		return wrapper.ResponseFailed(http.StatusBadRequest, contract.CreateStatusCode("PLAYER_NOT_IN_MATCH"), "Player is not in either team of this match", nil)
	}

	// Step 3: Create the goal
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

	// Step 4: Update match score based on which team scored
	if player.TeamID == match.HomeTeamID {
		match.HomeScore++
		l.Debug("incrementing home team score", zap.String("match_id", match.ID), zap.Int("new_score", match.HomeScore))
	} else {
		match.AwayScore++
		l.Debug("incrementing away team score", zap.String("match_id", match.ID), zap.Int("new_score", match.AwayScore))
	}

	// Step 5: Update match with new score and timestamp
	match.UpdatedAt = now

	if err := u.MatchRepository.Update(ctx, match); err != nil {
		l.Error("failed to update match score", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Goal created but failed to update match score", nil)
	}

	l.Debug("goal created and match score updated", zap.String("goal_id", g.ID), zap.String("match_id", match.ID), zap.Int("home_score", match.HomeScore), zap.Int("away_score", match.AwayScore))
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
