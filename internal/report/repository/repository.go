package repository

import (
	"context"
	"time"

	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
)

type Repository struct {
	DB database.IDBService
}

type IRepository interface {
	GoalsPerPlayer(context.Context) ([]GoalsPerPlayerRow, error)
	TeamGoals(context.Context) ([]TeamGoalsRow, error)
	MatchReport(context.Context) ([]MatchReportRow, error)
	GetMatchScorers(context.Context, string) ([]MatchScorerRow, error)
	GetTeamCumulativeHomeWins(context.Context, string, time.Time) (int64, error)
	GetTeamCumulativeAwayWins(context.Context, string, time.Time) (int64, error)
}

type MatchReportRow struct {
	MatchID      string
	MatchDate    time.Time
	MatchTime    time.Time
	HomeTeamID   string
	HomeTeamName string
	HomeLogoURL  string
	AwayTeamID   string
	AwayTeamName string
	AwayLogoURL  string
	HomeScore    int
	AwayScore    int
	Status       string
}

type GoalsPerPlayerRow struct {
	PlayerID string
	Goals    int64
}

type TeamGoalsRow struct {
	TeamID string
	Goals  int64
}

type MatchScorerRow struct {
	PlayerID     string
	PlayerName   string
	TeamName     string
	Position     string
	ShirtNumber  int
	MinuteScored int
}

func NewRepository(db database.IDBService) IRepository {
	return &Repository{DB: db}
}

func (r *Repository) GoalsPerPlayer(ctx context.Context) ([]GoalsPerPlayerRow, error) {
	var rows []GoalsPerPlayerRow
	// count goals grouped by player_id
	tx := r.DB.GetTransaction(ctx).Model(&model.Goal{}).
		Select("player_id as player_id, count(*) as goals").
		Group("player_id")
	if err := tx.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) TeamGoals(ctx context.Context) ([]TeamGoalsRow, error) {
	var rows []TeamGoalsRow
	// count goals grouped by match -> player's team is not directly available here
	// use players table join to attribute goals to team
	tx := r.DB.GetTransaction(ctx).
		Table("goals").
		Select("players.team_id as team_id, count(goals.id) as goals").
		Joins("left join players on players.id = goals.player_id").
		Group("players.team_id")
	if err := tx.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) MatchReport(ctx context.Context) ([]MatchReportRow, error) {
	var rows []MatchReportRow
	tx := r.DB.GetTransaction(ctx).
		Table("matches").
		Select(`
			matches.id as match_id,
			matches.match_date,
			matches.match_time,
			matches.home_team_id,
			home_team.name as home_team_name,
			home_team.logo_url as home_logo_url,
			matches.away_team_id,
			away_team.name as away_team_name,
			away_team.logo_url as away_logo_url,
			matches.home_score,
			matches.away_score,
			matches.status
		`).
		Joins("LEFT JOIN teams as home_team ON matches.home_team_id = home_team.id").
		Joins("LEFT JOIN teams as away_team ON matches.away_team_id = away_team.id").
		Where("matches.status = ?", "finished").
		Order("matches.match_date DESC")

	if err := tx.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetMatchScorers(ctx context.Context, matchID string) ([]MatchScorerRow, error) {
	var rows []MatchScorerRow
	tx := r.DB.GetTransaction(ctx).
		Table("goals").
		Select(`
			players.id as player_id,
			players.name as player_name,
			teams.name as team_name,
			players.position,
			players.shirt_number,
			goals.minute_scored
		`).
		Joins("INNER JOIN players ON players.id = goals.player_id").
		Joins("INNER JOIN teams ON teams.id = players.team_id").
		Where("goals.match_id = ?", matchID).
		Order("goals.minute_scored ASC")

	if err := tx.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetTeamCumulativeHomeWins(ctx context.Context, teamID string, upToDate time.Time) (int64, error) {
	var count int64
	tx := r.DB.GetTransaction(ctx).
		Model(&model.Match{}).
		Where("home_team_id = ?", teamID).
		Where("status = ?", "finished").
		Where("match_date <= ?", upToDate).
		Where("home_score > away_score")

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetTeamCumulativeAwayWins(ctx context.Context, teamID string, upToDate time.Time) (int64, error) {
	var count int64
	tx := r.DB.GetTransaction(ctx).
		Model(&model.Match{}).
		Where("away_team_id = ?", teamID).
		Where("status = ?", "finished").
		Where("match_date <= ?", upToDate).
		Where("away_score > home_score")

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
