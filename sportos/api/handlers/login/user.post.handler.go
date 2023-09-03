package login

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/mail"
	"strings"
	"unicode"
)

type UserPostHandler struct {
	UserPostRequest
}

type UserPostRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	UserType string `json:"userType,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Sport    string `json:"sport,omitempty"`
	City     string `json:"city,omitempty"`
}

func (r UserPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r UserPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *UserPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.UserPostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *UserPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
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
	if r.Password == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Password is mandatory")
	}
	if !verifyPassword(r.Password) {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Password isn't strong enough")
	}
	if r.Name == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Name is mandatory")
	}
	if r.Sport == "" && r.UserType != "player" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Sport is mandatory")
	}
	if _, err := DR.GetSportByName(r.Sport); r.Sport != "" && err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Sport " + r.Sport + " doesn't exist")
	}
	if r.City == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("City is mandatory")
	}
	return nil
}

func (r *UserPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sha256Hasher := sha256.New()
	data, _ := json.Marshal(r.Password)
	sha256Hasher.Write(data)
	dataHash := base64.URLEncoding.EncodeToString(sha256Hasher.Sum(nil))
	user := DR.User{
		Username:      r.Username,
		Email:         r.Email,
		EmailVerified: rand.Intn(9) - 9,
		PasswordHash:  dataHash,
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
	host := "https://localhost:4200"
	message := "\nPlease verify your email address for sportos by clicking on link " + host + "/verify?verifyToken=" + r.Username + "___" + fmt.Sprint(-user.EmailVerified) + "___" + user.PasswordHash
	DA.SendMail(message, "Verify email", []string{r.Email})
	tx.Commit()
	resMap := make(map[string]interface{})
	resMap["body"] = "Please verify your email to complete you registration"
	return resMap, nil
}

func verifyPassword(s string) bool {
	letters := len(s)
	var number, upper, special, lower bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			lower = true
		default:
			//
		}
	}
	sevenOrMore := letters >= 7
	return sevenOrMore && upper && lower && (number || special)
}
