package login

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

type SocialUserPostHandler struct {
	SocialUserPostRequest
}

type SocialUserPostRequest struct {
	Username string `json:"username,omitempty"`
	UserType string `json:"userType,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Sport    string `json:"sport,omitempty"`
	City     string `json:"city,omitempty"`
}

func (r SocialUserPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r SocialUserPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *SocialUserPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.SocialUserPostRequest)
	if err == nil {
		r.Email = strings.Split(r.Username, "_")[0] + "_" + r.Email
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *SocialUserPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.Email == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Email is mandatory")
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithMessage("Email isn't valid")
	}
	_, err = Repo.UserCrud.GetByEmail(ctx, r.Email, nil)
	if err == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Email is already in use")
	}
	if r.Username == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Username is mandatory")
	}
	_, err = Repo.UserCrud.GetById(ctx, r.Username, nil)
	if err == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_UNIQUE_CONSTRAINT).WithMessage("Username already exists")
	}
	if !DR.UserType(strings.ToLower(r.UserType)).IsValid() {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage(r.UserType + " isn't valid user type")
	}
	if strings.ToLower(r.UserType) == string(DR.UT_ADMIN) {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("You can't make admin user")
	}
	if r.Name == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Name is mandatory")
	}
	if r.Sport == "" && r.UserType != "player" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Sport is mandatory")
	}
	if _, err := DR.GetSportByName(r.Sport); err != nil && r.UserType != "player" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Sport " + r.Sport + " doesn't exist")
	}
	if r.City == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("City is mandatory")
	}
	return nil
}

func (r *SocialUserPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	user := DR.User{
		Username:      r.Username,
		Email:         r.Email,
		EmailVerified: 1,
		PasswordHash:  "",
		UserType:      DR.UserType(r.UserType),
	}
	tx, err := Repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	defer tx.Rollback()
	_, err = Repo.UserCrud.Create(ctx, user, tx, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	switch r.UserType {
	case string(DR.UT_PLAYER):
		player := DR.Player{
			Username: r.Username,
			Name:     r.Name,
			City:     r.City,
		}
		_, err := Repo.PlayerCrud.Create(ctx, player, tx, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
	case string(DR.UT_COACH):
		coach := DR.Coach{
			Username: r.Username,
			Name:     r.Name,
			Sport:    r.Sport,
			City:     r.City,
		}
		_, err := Repo.CoachCrud.Create(ctx, coach, tx, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
	case string(DR.UT_PLACE):
		place := DR.Place{
			Username: r.Username,
			Name:     r.Name,
			Sport:    r.Sport,
			City:     r.City,
		}
		_, err := Repo.PlaceCrud.Create(ctx, place, tx, nil)
		if err != nil {
			return nil, DA.InternalServerError(err)
		}
	}
	tx.Commit()
	resMap := make(map[string]interface{})
	resMap["body"] = struct{}{}
	return resMap, nil
}
