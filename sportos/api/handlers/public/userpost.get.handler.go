package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type UserpostGetHandler struct {
	UserId    *string `json:"userId,omitempty"`
	NotUserId *string `json:"notUserId,omitempty"`
}

func (r UserpostGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r UserpostGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *UserpostGetHandler) Init(httpReq *http.Request) DA.Error {
	r.UserId = DA.GetParameterFromURLQuery(httpReq, "userId")
	r.NotUserId = DA.GetParameterFromURLQuery(httpReq, "notUserId")
	return nil
}

func (r *UserpostGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *UserpostGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sp := DR.UserPostSearchParams{
		UserId:    r.UserId,
		NotUserId: r.NotUserId,
	}
	ret, err := Repo.UserPostsCrud.Search(ctx, sp, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	result := []DA.UserPost{}
	for _, post := range ret {
		name := ""
		player, err := Repo.PlayerCrud.GetById(ctx, post.UserId, nil)
		if err != nil {
			place, err := Repo.PlaceCrud.GetById(ctx, post.UserId, nil)
			if err != nil {
				coach, _ := Repo.CoachCrud.GetById(ctx, post.UserId, nil)
				name = coach.Name
			} else {
				name = place.Name
			}
		} else {
			name = player.Name
		}
		result = append(result, DA.UserPost{
			Name:       name,
			UserText:   post.UserText,
			ImageNames: post.ImageNames,
			CreatedAt:  post.CreatedAt,
		})
	}
	resMap := make(map[string]interface{})
	resMap["body"] = result
	return resMap, nil
}
