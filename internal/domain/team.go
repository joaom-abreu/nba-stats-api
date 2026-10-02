package domain

type Team struct {
	ID           int64  `json:"id"`
	Source       string `json:"source"`
	ExternalID   string `json:"external_id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Conference   string `json:"conference"`
}

type TeamFilter struct {
	Source     string
	Conference string
	Limit      int
	Offset     int
}
