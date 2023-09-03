package public

import (
	H "backend/internal/helpers"
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
	"strings"
)

type MatchGetHandler struct {
	PlayerId *string  `json:"playerId,omitempty"`
	Sports   []string `json:"sports"`
}

func (r MatchGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r MatchGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *MatchGetHandler) Init(httpReq *http.Request) DA.Error {
	r.PlayerId = DA.GetParameterFromURLQuery(httpReq, "playerId")
	r.Sports = DA.ParseCommaSeparated(DA.GetParameterFromURLQuery(httpReq, "sports"))
	return nil
}

func (r *MatchGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.PlayerId == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Player is mandatory")
	}
	if _, err := Repo.PlayerCrud.GetById(ctx, *r.PlayerId, nil); err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Player with username " + *r.PlayerId + " doesn't exist")
	}
	return nil
}

func (r *MatchGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	player, _ := Repo.PlayerCrud.GetById(ctx, *r.PlayerId, nil)
	sp := DR.MatchSearchParams{
		PlaceSearchParams: &DR.PlaceSearchParams{
			City: &player.City,
		},
		Sports: r.Sports,
		MatchSortParams: DR.MatchSortParams{
			EditInfoCUDSortParams: DR.EditInfoCUDSortParams{
				EditInfoCSortParams: DR.EditInfoCSortParams{
					CreatedAt: &DR.SortColumn{
						Prefix:    "ma",
						Column:    "created_at",
						Order:     0,
						Direction: 0,
					},
				},
			},
		},
	}
	matches, err := Repo.MatchCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	matches = r.formatMatches(ctx, Repo, matches)
	resMap := make(map[string]interface{})
	resMap["body"] = matches
	return resMap, nil
}

func (r *MatchGetHandler) formatMatches(ctx context.Context, Repo *crud.Repo, matches []DR.Match) []DR.Match {
	ret := []DR.Match{}
	for i := range matches {
		place, _ := Repo.PlaceCrud.GetById(ctx, matches[i].PlaceId, nil)
		matches[i].PlaceId = place.Name
		players := strings.Split(*matches[i].Players, ",")
		playerNames := ""
		for j := range players {
			player, _ := Repo.PlayerCrud.GetById(ctx, players[j], nil)
			playerNames = playerNames + player.Name + ","
		}
		playerNames = playerNames[0 : len(playerNames)-1]
		matches[i].PlayerNames = &playerNames
		if len(matches[i].Teams) != 0 {
			team := strings.Split(matches[i].Teams[0], ",")
			teamNames := ""
			for j := range team {
				player, _ := Repo.PlayerCrud.GetById(ctx, team[j], nil)
				teamNames = teamNames + player.Name + ","
			}
			teamNames = teamNames[0 : len(teamNames)-1]
			matches[i].Teams[0] = teamNames
			//
			team = strings.Split(matches[i].Teams[1], ",")
			teamNames = ""
			for j := range team {
				player, _ := Repo.PlayerCrud.GetById(ctx, team[j], nil)
				teamNames = teamNames + player.Name + ","
			}
			teamNames = teamNames[0 : len(teamNames)-1]
			matches[i].Teams[1] = teamNames
		}
		if !(matches[i].Status == DR.MS_FINISHED || (matches[i].Status == DR.MS_FULL && !H.Contains(players, *r.PlayerId))) {
			ret = append(ret, matches[i])
		}
	}
	return ret
}
