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

type MatchPostHandler struct {
	MatchPostRequest
}

type MatchPostRequest struct {
	StartTime *time.Time `json:"startTime,omitempty"`
	PlaceId   string     `json:"placeId,omitempty"`
	Players   string     `json:"players,omitempty"`
	Sport     string     `json:"sport,omitempty"`
}

type MatchPostResponse struct {
}

func (r MatchPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r MatchPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *MatchPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.MatchPostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *MatchPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.StartTime == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Start time is mandatory")
	}
	if r.Players == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Players are mandatory")
	}
	place, err := Repo.PlaceCrud.GetById(ctx, r.PlaceId, nil)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Place doesn't exist")
	}
	endTime := r.StartTime.Add(time.Hour)
	if place.Booking != nil {
		for _, booking := range *place.Booking {
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

func (r *MatchPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	match := DR.Match{
		Players:   &r.Players,
		PlaceId:   r.PlaceId,
		StartTime: r.StartTime,
		Status:    DR.MS_CREATED,
		Sport:     r.Sport,
	}
	tx, err := Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	defer tx.Rollback()
	ret, err := Repo.MatchCrud.Create(ctx, match, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	place, _ := Repo.PlaceCrud.GetById(ctx, r.PlaceId, tx)
	up := DR.PlaceUpdateParams{
		Id: r.PlaceId,
	}
	booking := place.Booking
	if booking == nil {
		booking = &DR.Booking{}
	}
	*booking = append(*booking, DR.Apointment{StartTime: *r.StartTime, EndTime: r.StartTime.Add(time.Hour)})
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
