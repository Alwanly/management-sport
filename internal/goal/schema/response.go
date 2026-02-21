package schema

type ResponseGoalCreate struct {
	ID string `json:"id"`
}

type ResponseGoalGet struct {
	ID           string `json:"id"`
	MatchID      string `json:"match_id"`
	PlayerID     string `json:"player_id"`
	MinuteScored int    `json:"minute_scored"`
}

type ResponseGoalItem struct {
	ID           string `json:"id"`
	MatchID      string `json:"match_id"`
	PlayerID     string `json:"player_id"`
	MinuteScored int    `json:"minute_scored"`
}

type ResponseGoalDelete struct{}
