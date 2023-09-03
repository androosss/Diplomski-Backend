package public

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"fmt"
	"net/http"
	"time"
)

type TimesGetHandler struct {
	PlaceId *string   `json:"username,omitempty"`
	Date    time.Time `json:"date,omitempty"`
}

type TimesGetResponse struct {
	Value string `json:"value"`
}

func (r TimesGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r TimesGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_CL}
}

func (r *TimesGetHandler) Init(httpReq *http.Request) DA.Error {
	r.PlaceId = DA.GetParameterFromURLQuery(httpReq, "username")
	var err error
	r.Date, err = time.Parse(time.RFC3339, *DA.GetParameterFromURLQuery(httpReq, "date"))
	if err != nil {
		return DA.ErrorBadRequest().WithMessage("Bad date format")
	}
	return nil
}

func (r *TimesGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.Date.Before(time.Now()) {
		return DA.ErrorBadRequest().WithMessage("Date must be after now")
	}
	if r.PlaceId == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Place is mandatory for creating match")
	}
	if _, err := Repo.PlaceCrud.GetById(ctx, *r.PlaceId, nil); err != nil {
		_, err := Repo.CoachCrud.GetById(ctx, *r.PlaceId, nil)
		if err != nil {
			return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("User with username " + *r.PlaceId + " doesn't exist")
		}
	}
	return nil
}

func (r *TimesGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	ret := []TimesGetResponse{}
	var bookings *DR.Booking
	place, err := Repo.PlaceCrud.GetById(ctx, *r.PlaceId, nil)
	if err != nil {
		coach, _ := Repo.CoachCrud.GetById(ctx, *r.PlaceId, nil)
		bookings = coach.Booking
	} else {
		bookings = place.Booking
	}

	for i := 0; i < 24; i++ {
		tempStart := r.Date.Add(time.Hour * time.Duration(i))
		tempEnd := r.Date.Add(time.Hour * time.Duration(i+1))
		toAppend := true
		if bookings != nil {
			for _, booking := range *bookings {
				if booking.PracticeId != "" && booking.Accepted {
					continue
				}
				if !(DA.AfterEqual(tempStart, booking.EndTime) || DA.BeforeEqual(tempEnd, booking.StartTime)) {
					toAppend = false
				}
			}
		}
		if toAppend {
			ret = append(ret, TimesGetResponse{Value: toTimeStr(i)})
		}
	}
	resMap := make(map[string]interface{})
	resMap["body"] = ret
	return resMap, nil
}

func toTimeStr(i int) string {
	if i < 10 {
		return "0" + fmt.Sprint(i) + ":00"
	}
	return fmt.Sprint(i) + ":00"
}
