package dto

import "fmt"

type Sport struct {
	Name      string   `json:"name,omitempty"`
	TeamSize  int      `json:"teamSize,omitempty"`
	Positions []string `json:"positions,omitempty"`
}

var Sports []Sport = []Sport{
	{
		Name:     "Tennis",
		TeamSize: 1,
	},
	{
		Name:     "Table tennis",
		TeamSize: 1,
	},
	{
		Name:     "Basketball",
		TeamSize: 5,
		Positions: []string{
			"Point guard",
			"Shooting guard",
			"Small forward",
			"Power forward",
			"Center",
		},
	},
	{
		Name:     "Volleyball",
		TeamSize: 6,
		Positions: []string{
			"Outside hitter",
			"Opposite",
			"Setter",
			"Middle blocker",
			"Libero",
		},
	},
	{
		Name:     "Handball",
		TeamSize: 7,
		Positions: []string{
			"Goalkeeper",
			"Left wing",
			"Left back",
			"Middle back",
			"Line player",
			"Right back",
			"Right wing",
		},
	},
	{
		Name:     "Football",
		TeamSize: 11,
		Positions: []string{
			"Attack",
			"Middle field",
			"Defence",
			"Goalkeeper",
		},
	},
	{
		Name:     "Swimming",
		TeamSize: 1,
	},
	{
		Name:     "Fitness",
		TeamSize: 1,
	},
	{
		Name:     "Bodybuilding",
		TeamSize: 1,
	},
}

func GetSportByName(s string) (Sport, error) {
	for _, sport := range Sports {
		if sport.Name == s {
			return sport, nil
		}
	}
	return Sport{}, fmt.Errorf("sport %s doesn't exist", s)
}
