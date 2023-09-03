package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type TeamsPatchHandler struct {
	TeamsPatchRequest
	Players string
	sport   string
}

type TeamsPatchRequest struct {
	Id          string  `json:"id,omitempty"`
	PlayerToAdd *string `json:"player,omitempty"`
}

func (r TeamsPatchHandler) SupportedMethod() string {
	return http.MethodPatch
}

func (r TeamsPatchHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TeamsPatchHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.TeamsPatchRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *TeamsPatchHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if team, err := Repo.TeamCrud.GetById(ctx, r.Id, nil); err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Match with id " + r.Id + " doesn't exist")
	} else {
		if r.PlayerToAdd != nil && strings.Contains(team.Players, *r.PlayerToAdd) {
			return DA.ErrorBadRequest().WithMessage("Player is already in that team")
		}
		if r.PlayerToAdd != nil && team.Status == DR.TS_FULL {
			return DA.ErrorBadRequest().WithMessage("Team is full already")
		}
		r.sport = team.Sport
		r.Players = team.Players
	}
	return nil
}

func (r *TeamsPatchHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sport, err := DR.GetSportByName(r.sport)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	var status *DR.TeamStatus
	if r.PlayerToAdd != nil {
		r.Players += "," + *r.PlayerToAdd
	}
	if len(strings.Split(r.Players, ",")) == sport.TeamSize {
		status = new(DR.TeamStatus)
		*status = DR.TS_FULL
	}
	up := DR.TeamUpdateParams{
		Id:      r.Id,
		Players: &r.Players,
		Status:  status,
	}
	ret, err := Repo.TeamCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
