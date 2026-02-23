package schema

type ResponsePlayerCreate struct {
	ID string `json:"id"`
}

type ResponsePlayerGet struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"`
	Name        string `json:"name"`
	HeightCM    int    `json:"height_cm"`
	WeightKG    int    `json:"weight_kg"`
	Position    string `json:"position"`
	ShirtNumber int    `json:"shirt_number"`
	TeamName    string `json:"team_name"`
}

type ResponsePlayerItem struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	ShirtNumber int    `json:"shirt_number"`
	TeamName    string `json:"team_name"`
}

type ResponsePlayerUpdate struct {
	ID string `json:"id"`
}

type ResponsePlayerDelete struct{}
