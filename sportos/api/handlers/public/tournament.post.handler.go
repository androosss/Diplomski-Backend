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

type TournamentPostHandler struct {
	TournamentPostRequest
}

type TournamentPostRequest struct {
	StartTime *time.Time `json:"startTime,omitempty"`
	placeId   string
	sport     string
	Name      string `json:"name,omitempty"`
}

type TournamentPostResponse struct {
}

func (r TournamentPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r TournamentPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TournamentPostHandler) Init(httpReq *http.Request) DA.Error {
	r.placeId = DA.GetUserIdFromContext(httpReq.Context())
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.TournamentPostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *TournamentPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.StartTime == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Start time is mandatory")
	}
	place, err := Repo.PlaceCrud.GetById(ctx, r.placeId, nil)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Place doesn't exist")
	}
	events, err := Repo.EventCrud.Search(ctx, DR.EventSearchParams{Name: &r.Name}, nil)
	if err != nil {
		return DA.InternalServerError(err)
	}
	if len(events) != 0 {
		return DA.ErrorBadRequest().WithMessage("Event with same name alredy exists")
	}
	year, month, day := r.StartTime.Date()
	startTime := time.Date(year, month, day, 0, 0, 0, 0, r.StartTime.Location())
	if startTime.Before(time.Now()) {
		return DA.ErrorBadRequest().WithMessage("Event must be in the future")
	}
	endTime := time.Date(year, month, day, 0, 0, 0, 0, r.StartTime.Location()).AddDate(0, 0, 1)
	if place.Booking != nil {
		for _, booking := range *place.Booking {
			if (DA.AfterEqual(startTime, booking.StartTime) && !DA.AfterEqual(startTime, booking.EndTime)) ||
				(DA.BeforeEqual(startTime, booking.StartTime) && !DA.BeforeEqual(endTime, booking.EndTime)) {
				return DA.ErrorBadRequest().WithMessage("That day is already occupied")
			}
		}
	}
	r.sport = place.Sport
	return nil
}

func (r *TournamentPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	tx, err := Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	defer tx.Rollback()
	tournament := DR.Event{
		Owner:  r.placeId,
		Time:   r.StartTime,
		Status: DR.ES_CREATED,
		Sport:  r.sport,
		Name:   r.Name,
	}
	ret, err := Repo.EventCrud.Create(ctx, tournament, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	place, _ := Repo.PlaceCrud.GetById(ctx, r.placeId, tx)
	up := DR.PlaceUpdateParams{
		Id: r.placeId,
	}
	booking := place.Booking
	if booking == nil {
		booking = &DR.Booking{}
	}
	year, month, day := r.StartTime.Date()
	startTime := time.Date(year, month, day, 0, 0, 0, 0, r.StartTime.Location())
	endTime := time.Date(year, month, day, 0, 0, 0, 0, r.StartTime.Location()).AddDate(0, 0, 1)
	*booking = append(*booking, DR.Apointment{StartTime: startTime, EndTime: endTime})
	up.Booking = booking
	_, err = Repo.PlaceCrud.Update(ctx, up, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	tx.Commit()
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}
