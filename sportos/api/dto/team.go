package dto

type Team struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Sport   string   `json:"sport"`
	Players []string `json:"players"`
}
