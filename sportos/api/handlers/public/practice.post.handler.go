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

type PracticePostHandler struct {
	PracticePostRequest
}

type PracticePostRequest struct {
	StartTime *time.Time `json:"startTime,omitempty"`
	userId    string
	CoachId   string `json:"coachId,omitempty"`
	Sport     string `json:"sport,omitempty"`
}

func (r PracticePostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r PracticePostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *PracticePostHandler) Init(httpReq *http.Request) DA.Error {
	r.userId = DA.GetUserIdFromContext(httpReq.Context())
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.PracticePostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *PracticePostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.StartTime == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Start time is mandatory")
	}
	coach, err := Repo.CoachCrud.GetById(ctx, r.CoachId, nil)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Coach doesn't exist")
	}
	endTime := r.StartTime.Add(time.Hour)
	if coach.Booking != nil {
		for _, booking := range *coach.Booking {
			if !booking.Accepted {
				continue
			}
			if !(DA.AfterEqual(*r.StartTime, booking.EndTime) || DA.BeforeEqual(endTime, booking.StartTime)) {
				return DA.ErrorBadRequest().WithMessage("That appointment is already occupied")
			}
		}
	}
	if _, err := DR.GetSportByName(r.Sport); err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Sport doesn't exist")
	}
	return nil
}

func (r *PracticePostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	practice := DR.Practice{
		PlayerId:  r.userId,
		CoachId:   r.CoachId,
		StartTime: r.StartTime,
		Status:    DR.PS_CREATED,
		Sport:     r.Sport,
	}
	tx, err := Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	defer tx.Rollback()
	ret, err := Repo.PracticeCrud.Create(ctx, practice, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	place, _ := Repo.CoachCrud.GetById(ctx, r.CoachId, tx)
	up := DR.CoachUpdateParams{
		Id: r.CoachId,
	}
	booking := place.Booking
	if booking == nil {
		booking = &DR.Booking{}
	}
	*booking = append(*booking, DR.Apointment{StartTime: *r.StartTime, EndTime: r.StartTime.Add(time.Hour), Accepted: false, PracticeId: ret.PracticeId})
	up.Booking = booking
	_, err = Repo.CoachCrud.Update(ctx, up, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	tx.Commit()
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
