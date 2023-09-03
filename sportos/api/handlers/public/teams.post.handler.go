package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
)

type TeamsPostHandler struct {
	TeamsPostRequest
}

type TeamsPostRequest struct {
	Name     string `json:"name"`
	Sport    string `json:"sport,omitempty"`
	userId   string
	teamSize int
}

type TeamsPostResponse struct {
}

func (r TeamsPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r TeamsPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TeamsPostHandler) Init(httpReq *http.Request) DA.Error {
	r.userId = DA.GetUserIdFromContext(httpReq.Context())
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.TeamsPostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *TeamsPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if sport, err := DR.GetSportByName(r.Sport); err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Sport doesn't exist")
	} else {
		r.teamSize = sport.TeamSize
	}
	if Repo.TeamCrud.CheckConstraints(ctx, DR.Team{Name: r.Name, Sport: r.Sport}, nil) {
		return DA.ErrorBadRequest().WithMessage("Team with that name already exists for that sport")
	}
	return nil
}

func (r *TeamsPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	team := DR.Team{
		Status:  DR.TS_CREATED,
		Sport:   r.Sport,
		Players: r.userId,
		Name:    r.Name,
	}
	if r.teamSize == 1 {
		team.Status = DR.TS_FULL
	}
	ret, err := Repo.TeamCrud.Create(ctx, team, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
