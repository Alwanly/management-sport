package usecase

import (
	"context"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/report/repository"
	"github.com/Alwanly/management-sport/internal/report/schema"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	GoalsPerPlayer(context.Context) wrapper.JSONResult
	TeamGoals(context.Context) wrapper.JSONResult
	MatchReport(context.Context) wrapper.JSONResult
}

func NewUseCase(u UseCase) IUseCase {
	return &UseCase{Config: u.Config, Logger: u.Logger, Repository: u.Repository}
}

func (u *UseCase) GoalsPerPlayer(ctx context.Context) wrapper.JSONResult {
	rows, err := u.Repository.GoalsPerPlayer(ctx)
	if err != nil {
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "failed to get report", nil)
	}
	resp := make([]schema.ResponseGoalsPerPlayerItem, len(rows))
	for i, r := range rows {
		resp[i] = schema.ResponseGoalsPerPlayerItem{PlayerID: r.PlayerID, Goals: r.Goals}
	}
	return wrapper.ResponseSuccess(200, resp)
}

func (u *UseCase) TeamGoals(ctx context.Context) wrapper.JSONResult {
	rows, err := u.Repository.TeamGoals(ctx)
	if err != nil {
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "failed to get report", nil)
	}
	resp := make([]schema.ResponseTeamGoalsItem, len(rows))
	for i, r := range rows {
		resp[i] = schema.ResponseTeamGoalsItem{TeamID: r.TeamID, Goals: r.Goals}
	}
	return wrapper.ResponseSuccess(200, resp)
}

func (u *UseCase) MatchReport(ctx context.Context) wrapper.JSONResult {
	rows, err := u.Repository.MatchReport(ctx)
	if err != nil {
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "failed to get match report", nil)
	}

	resp := make([]schema.ResponseMatchReportItem, len(rows))
	for i, r := range rows {
		result := "Draw"
		if r.HomeScore > r.AwayScore {
			result = "Home Win"
		} else if r.HomeScore < r.AwayScore {
			result = "Away Win"
		}

		resp[i] = schema.ResponseMatchReportItem{
			MatchID:   r.MatchID,
			MatchDate: r.MatchDate,
			MatchTime: r.MatchTime,
			HomeTeam: schema.MatchTeamInfo{
				ID:      r.HomeTeamID,
				Name:    r.HomeTeamName,
				LogoURL: r.HomeLogoURL,
			},
			AwayTeam: schema.MatchTeamInfo{
				ID:      r.AwayTeamID,
				Name:    r.AwayTeamName,
				LogoURL: r.AwayLogoURL,
			},
			FinalScore: schema.MatchScoreInfo{
				Home: r.HomeScore,
				Away: r.AwayScore,
			},
			Result: result,
			Status: r.Status,
		}
	}

	return wrapper.ResponseSuccess(200, resp)
}
