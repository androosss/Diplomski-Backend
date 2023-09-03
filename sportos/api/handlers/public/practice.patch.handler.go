package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type PracticePatchHandler struct {
	PracticePatchRequest
	coachId string
}

type PracticePatchRequest struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status"`
}

func (r PracticePatchHandler) SupportedMethod() string {
	return http.MethodPatch
}

func (r PracticePatchHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *PracticePatchHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.PracticePatchRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *PracticePatchHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if practice, err := Repo.PracticeCrud.GetById(ctx, r.Id, nil); err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Match with id " + r.Id + " doesn't exist")
	} else {
		r.coachId = practice.CoachId
	}
	return nil
}

func (r *PracticePatchHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	up := DR.PracticeUpdateParams{
		Id:     r.Id,
		Status: (*DR.PracticeStatus)(&r.Status),
	}
	if r.Status == string(DR.PS_ACCEPTED) {
		coach, err := Repo.CoachCrud.GetById(ctx, r.coachId, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
		var start, end time.Time
		for i := range *coach.Booking {
			if (*coach.Booking)[i].PracticeId == r.Id {
				(*coach.Booking)[i].Accepted = true
				start = (*coach.Booking)[i].StartTime
				end = (*coach.Booking)[i].EndTime
			}
		}
		newBooking := DR.Booking{}
		for _, booking := range *coach.Booking {
			if booking.Accepted || DA.AfterEqual(start, booking.EndTime) || DA.BeforeEqual(end, booking.StartTime) {
				newBooking = append(newBooking, booking)
			} else {
				denied := DR.PS_DENIED
				up := DR.PracticeUpdateParams{
					Id:     booking.PracticeId,
					Status: &denied,
				}
				_, err := Repo.PracticeCrud.Update(ctx, up, nil, nil)
				if err != nil {
					return nil, DA.InternalServerError(err)
				}
			}
		}
		*coach.Booking = newBooking
		up := DR.CoachUpdateParams{
			Id:      coach.Username,
			Booking: coach.Booking,
		}
		_, err = Repo.CoachCrud.Update(ctx, up, nil, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
	}
	ret, err := Repo.PracticeCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
