package dto

import (
	"time"
)

type Statistic struct {
	MyTeam  []string  `json:"myTeam,omitempty"`
	OppTeam []string  `json:"oppTeam,omitempty"`
	Date    time.Time `json:"date,omitempty"`
	Score   string    `json:"score,omitempty"`
}

type Statistics struct {
	Matches     []Statistic  `json:"matches,omitempty"`
	WinRatio    string       `json:"winRatio,omitempty"`
	Tournaments []Tournament `json:"tournaments,omitempty"`
}

type Tournament struct {
	MyTeam     []string `json:"myTeam,omitempty"`
	Ranking    int      `json:"ranking"`
	Tournament string   `json:"tournament"`
}

type StatMap map[string]Statistics
