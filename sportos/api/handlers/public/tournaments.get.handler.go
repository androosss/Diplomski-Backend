package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
	"strings"
)

type TournamentsGetHandler struct {
	sport  *string
	place  *string
	status *string
}

func (r TournamentsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r TournamentsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TournamentsGetHandler) Init(httpReq *http.Request) DA.Error {
	r.sport = DA.GetParameterFromURLQuery(httpReq, "sports")
	r.place = DA.GetParameterFromURLQuery(httpReq, "place")
	r.status = DA.GetParameterFromURLQuery(httpReq, "status")
	return nil
}

func (r *TournamentsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *TournamentsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sp := DR.EventSearchParams{
		Owner:  r.place,
		Status: r.status,
	}
	if r.sport != nil {
		sp.Sports = strings.Split(*r.sport, ",")
	}
	events, err := Repo.EventCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = events
	return resMap, nil
}
