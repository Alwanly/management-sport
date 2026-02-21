package schema

type ResponseGoalsPerPlayerItem struct {
	PlayerID string `json:"player_id"`
	Goals    int64  `json:"goals"`
}

type ResponseTeamGoalsItem struct {
	TeamID string `json:"team_id"`
	Goals  int64  `json:"goals"`
}
