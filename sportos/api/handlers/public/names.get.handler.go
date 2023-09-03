package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type NameGetHandler struct {
	id string
}

func (r NameGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r NameGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *NameGetHandler) Init(httpReq *http.Request) DA.Error {
	r.id = mux.Vars(httpReq)["id"]
	return nil
}

func (r *NameGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.id == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("id is mandatory")
	}
	return nil
}

func (r *NameGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	name, err := Repo.GetNameForId(ctx, r.id)
	if err != nil {
		return nil, DA.ErrorBadRequest().WithMessage("that id does not exist")
	}
	resMap := make(map[string]interface{})
	resMap["body"] = name
	return resMap, nil
}
