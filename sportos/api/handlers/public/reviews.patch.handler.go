package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
)

type ReviewsPatchHandler struct {
	ReviewsPatchRequest
	userId  string
	isCoach bool
}

type ReviewsPatchRequest struct {
	Id      string  `json:"id,omitempty"`
	Comment string  `json:"comment,omitempty"`
	Grade   float64 `json:"grade,omitempty"`
}

func (r ReviewsPatchHandler) SupportedMethod() string {
	return http.MethodPatch
}

func (r ReviewsPatchHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *ReviewsPatchHandler) Init(httpReq *http.Request) DA.Error {
	r.userId = DA.GetUserIdFromContext(httpReq.Context())
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.ReviewsPatchRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *ReviewsPatchHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if _, err := Repo.CoachCrud.GetById(ctx, r.Id, nil); err == nil {
		r.isCoach = true
	} else {
		if _, err := Repo.PlaceCrud.GetById(ctx, r.Id, nil); err != nil {
			return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("User with id " + r.Id + " doesn't exist")
		}
	}
	return nil
}

func (r *ReviewsPatchHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	resMap := make(map[string]interface{})
	if r.isCoach {
		coach, _ := Repo.CoachCrud.GetById(ctx, r.Id, nil)
		if coach.Reviews == nil {
			coach.Reviews = &DR.Reviews{}
		}
		isUpdated := false
		sum := 0.0
		for i := range coach.Reviews.Reviews {
			if coach.Reviews.Reviews[i].UserId == r.userId {
				coach.Reviews.Reviews[i].Grade = r.Grade
				coach.Reviews.Reviews[i].Comment = r.Comment
				isUpdated = true
			}
			sum += coach.Reviews.Reviews[i].Grade
		}
		if !isUpdated {
			coach.Reviews.Reviews = append(coach.Reviews.Reviews, DR.Review{
				UserId:  r.userId,
				Grade:   r.Grade,
				Comment: r.Comment,
			})
			sum += r.Grade
		}
		if len(coach.Reviews.Reviews) != 0 {
			coach.Reviews.Average = sum / float64(len(coach.Reviews.Reviews))
		}
		up := DR.CoachUpdateParams{
			Id:      r.Id,
			Reviews: coach.Reviews,
		}
		coach, err := Repo.CoachCrud.Update(ctx, up, nil, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
		resMap["body"] = coach
	} else {
		place, _ := Repo.PlaceCrud.GetById(ctx, r.Id, nil)
		if place.Reviews == nil {
			place.Reviews = &DR.Reviews{}
		}
		isUpdated := false
		sum := 0.0
		for i := range place.Reviews.Reviews {
			if place.Reviews.Reviews[i].UserId == r.userId {
				place.Reviews.Reviews[i].Grade = r.Grade
				place.Reviews.Reviews[i].Comment = r.Comment
			}
			sum += place.Reviews.Reviews[i].Grade
		}
		if !isUpdated {
			place.Reviews.Reviews = append(place.Reviews.Reviews, DR.Review{
				UserId:  r.userId,
				Grade:   r.Grade,
				Comment: r.Comment,
			})
			sum += r.Grade
		}
		if len(place.Reviews.Reviews) != 0 {
			place.Reviews.Average = sum / float64(len(place.Reviews.Reviews))
		}
		up := DR.PlaceUpdateParams{
			Id:      r.Id,
			Reviews: place.Reviews,
		}
		place, err := Repo.PlaceCrud.Update(ctx, up, nil, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
		resMap["body"] = place
	}
	return resMap, nil
}
