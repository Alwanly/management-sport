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
	l := u.Logger.With(zap.String("usecase", "MatchReport"))

	// Get base match data
	rows, err := u.Repository.MatchReport(ctx)
	if err != nil {
		l.Error("failed to get match report", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "failed to get match report", nil)
	}

	// Build matches array
	matches := make([]schema.ResponseMatchReportItem, len(rows))
	scorersMap := make(map[string][]schema.ResponseMatchScorer)
	teamStatsMap := make(map[string]schema.ResponseTeamStatistics)
	processedTeams := make(map[string]bool)

	for i, r := range rows {
		// Calculate result
		result := "Draw"
		if r.HomeScore > r.AwayScore {
			result = "Home Win"
		} else if r.HomeScore < r.AwayScore {
			result = "Away Win"
		}

		// Build match item
		matches[i] = schema.ResponseMatchReportItem{
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

		// Get match scorers
		scorers, err := u.Repository.GetMatchScorers(ctx, r.MatchID)
		if err != nil {
			l.Error("failed to get match scorers", zap.String("match_id", r.MatchID), zap.Error(err))
			// Continue without scorers for this match
			scorersMap[r.MatchID] = []schema.ResponseMatchScorer{}
		} else {
			scorerItems := make([]schema.ResponseMatchScorer, len(scorers))
			for j, s := range scorers {
				scorerItems[j] = schema.ResponseMatchScorer{
					PlayerID:     s.PlayerID,
					PlayerName:   s.PlayerName,
					TeamName:     s.TeamName,
					Position:     s.Position,
					ShirtNumber:  s.ShirtNumber,
					MinuteScored: s.MinuteScored,
				}
			}
			scorersMap[r.MatchID] = scorerItems
		}

		// Process home team statistics
		if !processedTeams[r.HomeTeamID] {
			homeWins, err := u.Repository.GetTeamCumulativeHomeWins(ctx, r.HomeTeamID, r.MatchDate)
			if err != nil {
				l.Error("failed to get home wins", zap.String("team_id", r.HomeTeamID), zap.Error(err))
				homeWins = 0
			}
			awayWins, err := u.Repository.GetTeamCumulativeAwayWins(ctx, r.HomeTeamID, r.MatchDate)
			if err != nil {
				l.Error("failed to get away wins", zap.String("team_id", r.HomeTeamID), zap.Error(err))
				awayWins = 0
			}
			teamStatsMap[r.HomeTeamID] = schema.ResponseTeamStatistics{
				TeamID:             r.HomeTeamID,
				TeamName:           r.HomeTeamName,
				CumulativeHomeWins: homeWins,
				CumulativeAwayWins: awayWins,
			}
			processedTeams[r.HomeTeamID] = true
		}

		// Process away team statistics
		if !processedTeams[r.AwayTeamID] {
			homeWins, err := u.Repository.GetTeamCumulativeHomeWins(ctx, r.AwayTeamID, r.MatchDate)
			if err != nil {
				l.Error("failed to get home wins", zap.String("team_id", r.AwayTeamID), zap.Error(err))
				homeWins = 0
			}
			awayWins, err := u.Repository.GetTeamCumulativeAwayWins(ctx, r.AwayTeamID, r.MatchDate)
			if err != nil {
				l.Error("failed to get away wins", zap.String("team_id", r.AwayTeamID), zap.Error(err))
				awayWins = 0
			}
			teamStatsMap[r.AwayTeamID] = schema.ResponseTeamStatistics{
				TeamID:             r.AwayTeamID,
				TeamName:           r.AwayTeamName,
				CumulativeHomeWins: homeWins,
				CumulativeAwayWins: awayWins,
			}
			processedTeams[r.AwayTeamID] = true
		}
	}

	// Build enhanced response
	enhancedReport := schema.ResponseEnhancedMatchReport{
		Matches:        matches,
		Scorers:        scorersMap,
		TeamStatistics: teamStatsMap,
	}

	l.Debug("match report generated",
		zap.Int("matches_count", len(matches)),
		zap.Int("teams_count", len(teamStatsMap)))

	return wrapper.ResponseSuccess(200, enhancedReport)
}
