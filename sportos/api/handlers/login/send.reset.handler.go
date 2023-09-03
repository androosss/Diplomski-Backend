package login

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/mail"
)

type SendResetPostHandler struct {
	SendResetPostRequest
}

type SendResetPostRequest struct {
	Email string `json:"email,omitempty"`
}

func (r SendResetPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r SendResetPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *SendResetPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.SendResetPostRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *SendResetPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.Email == "" {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Email is mandatory")
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithMessage("Email isn't valid")
	}
	_, err = Repo.UserCrud.GetByEmail(ctx, r.Email, nil)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Email doesn't belong to any user")
	}
	return nil
}

func (r *SendResetPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	host := "https://localhost:4200"
	user, _ := Repo.UserCrud.GetByEmail(ctx, r.Email, nil)
	resetToken := base64.URLEncoding.EncodeToString([]byte(user.Username + "___" + user.PasswordHash))
	message := "\nReset your sportos password by clicking on link " + host + "/reset-password?resetToken=" + resetToken
	DA.SendMail(message, "Reset password", []string{r.Email})
	resMap := make(map[string]interface{})
	resMap["body"] = struct{}{}
	return resMap, nil
}
