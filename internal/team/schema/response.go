package schema

type ResponseTeamCreate struct {
	ID string `json:"id"`
}

type ResponseTeamGet struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	Address     string `json:"address"`
	City        string `json:"city"`
}

type ResponseTeamItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	City        string `json:"city"`
}

type ResponseTeamUpdate struct {
	ID string `json:"id"`
}

type ResponseTeamDelete struct{}
