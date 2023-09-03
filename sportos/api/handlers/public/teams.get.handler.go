package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
	"strings"
)

type TeamsGetHandler struct {
	skipUser *string
	sports   *string
	owner    *string
	status   *string
}

func (r TeamsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r TeamsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TeamsGetHandler) Init(httpReq *http.Request) DA.Error {
	r.skipUser = DA.GetParameterFromURLQuery(httpReq, "skipUser")
	r.sports = DA.GetParameterFromURLQuery(httpReq, "sports")
	r.owner = DA.GetParameterFromURLQuery(httpReq, "owner")
	r.status = DA.GetParameterFromURLQuery(httpReq, "status")
	return nil
}

func (r *TeamsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *TeamsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	ret := []DA.Team{}
	sp := DR.TeamSearchParams{
		Status:        r.status,
		UserNotInTeam: r.skipUser,
		Owner:         r.owner,
	}
	if r.sports != nil {
		sp.Sports = strings.Split(*r.sports, ",")
	}
	teams, err := Repo.TeamCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	for _, team := range teams {
		apiTeam := DA.Team{
			Id:    team.TeamId,
			Name:  team.Name,
			Sport: team.Sport,
		}
		for _, playerId := range strings.Split(team.Players, ",") {
			player, _ := Repo.PlayerCrud.GetById(ctx, playerId, nil)
			apiTeam.Players = append(apiTeam.Players, player.Name)
		}
		ret = append(ret, apiTeam)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
