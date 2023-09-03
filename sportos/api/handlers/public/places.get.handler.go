package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type PlacesGetHandler struct {
	PlaceId *string `json:"username,omitempty"`
	City    *string `json:"city,omitempty"`
	Name    *string `json:"name,omitempty"`
	Sport   *string `json:"sport,omitempty"`
}

func (r PlacesGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r PlacesGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *PlacesGetHandler) Init(httpReq *http.Request) DA.Error {
	r.PlaceId = DA.GetParameterFromURLQuery(httpReq, "playerId")
	r.City = DA.GetParameterFromURLQuery(httpReq, "city")
	r.Name = DA.GetParameterFromURLQuery(httpReq, "name")
	r.Sport = DA.GetParameterFromURLQuery(httpReq, "sport")
	return nil
}

func (r *PlacesGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *PlacesGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sp := DR.PlaceSearchParams{
		Name:     r.Name,
		City:     r.City,
		Username: r.PlaceId,
		Sport:    r.Sport,
	}
	places, err := Repo.PlaceCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = places
	return resMap, nil
}
