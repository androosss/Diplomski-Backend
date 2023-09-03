package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type ReviewsGetHandler struct {
	Id      *string
	isCoach bool
}

func (r ReviewsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r ReviewsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *ReviewsGetHandler) Init(httpReq *http.Request) DA.Error {
	r.Id = DA.GetParameterFromURLQuery(httpReq, "id")
	return nil
}

func (r *ReviewsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.Id == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Id is mandatory")
	}
	if _, err := Repo.CoachCrud.GetById(ctx, *r.Id, nil); err == nil {
		r.isCoach = true
	} else {
		if _, err := Repo.PlaceCrud.GetById(ctx, *r.Id, nil); err != nil {
			return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("User with id " + *r.Id + " doesn't exist")
		}
	}
	return nil
}

func (r *ReviewsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	ret := &DR.Reviews{}
	if r.isCoach {
		coach, _ := Repo.CoachCrud.GetById(ctx, *r.Id, nil)
		ret = coach.Reviews
	} else {
		place, _ := Repo.PlaceCrud.GetById(ctx, *r.Id, nil)
		ret = place.Reviews
	}
	resMap := make(map[string]interface{})
	resMap["body"] = apiReviews(ctx, Repo, ret)
	return resMap, nil
}

type ApiReviews struct {
	Average float64     `json:"average"`
	Reviews []ApiReview `json:"reviews"`
}

type ApiReview struct {
	Comment string  `json:"comment"`
	Rating  float64 `json:"rating"`
	UserId  string  `json:"userId"`
	Name    string  `json:"name"`
}

func apiReviews(ctx context.Context, repo *crud.Repo, reviews *DR.Reviews) ApiReviews {
	ret := ApiReviews{}
	if reviews == nil {
		return ret
	}
	ret.Average = reviews.Average
	for _, review := range reviews.Reviews {
		name, _ := repo.GetNameForId(ctx, review.UserId)
		ret.Reviews = append(ret.Reviews, ApiReview{
			Comment: review.Comment,
			UserId:  review.UserId,
			Rating:  review.Grade,
			Name:    name,
		})
	}
	return ret
}
