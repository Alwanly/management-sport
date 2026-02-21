package repository

import (
	"context"

	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
)

type Repository struct {
	DB database.IDBService
}

type IRepository interface {
	GoalsPerPlayer(context.Context) ([]GoalsPerPlayerRow, error)
	TeamGoals(context.Context) ([]TeamGoalsRow, error)
}

type GoalsPerPlayerRow struct {
	PlayerID string
	Goals    int64
}

type TeamGoalsRow struct {
	TeamID string
	Goals  int64
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
