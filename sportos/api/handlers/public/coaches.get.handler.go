package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type CoachsGetHandler struct {
	CoachId *string `json:"username,omitempty"`
	City    *string `json:"city,omitempty"`
	Name    *string `json:"name,omitempty"`
	Sport   *string `json:"sport,omitempty"`
}

func (r CoachsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r CoachsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *CoachsGetHandler) Init(httpReq *http.Request) DA.Error {
	r.CoachId = DA.GetParameterFromURLQuery(httpReq, "playerId")
	r.City = DA.GetParameterFromURLQuery(httpReq, "city")
	r.Name = DA.GetParameterFromURLQuery(httpReq, "name")
	r.Sport = DA.GetParameterFromURLQuery(httpReq, "sport")
	return nil
}

func (r *CoachsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *CoachsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sp := DR.CoachSearchParams{
		Name:     r.Name,
		City:     r.City,
		Username: r.CoachId,
		Sport:    r.Sport,
	}
	coachs, err := Repo.CoachCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = coachs
	return resMap, nil
}
