package login

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type UserVerifyPostHandler struct {
	VerifyToken string `json:"verifyToken,omitempty"`
}

func (r UserVerifyPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r UserVerifyPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *UserVerifyPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *UserVerifyPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *UserVerifyPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	decoded := strings.Split(r.VerifyToken, "___")
	if len(decoded) != 3 {
		return nil, DA.ErrorBadRequest()
	}
	username := decoded[0]
	salt := int(decoded[1][0] - '0')
	passwordHash := decoded[2]
	user, err := Repo.UserCrud.GetById(ctx, username, nil)
	if err != nil {
		return nil, DA.ErrorBadRequest()
	}
	if user.PasswordHash != passwordHash || user.EmailVerified != -salt {
		return nil, DA.ErrorBadRequest()
	}
	up := DR.UserUpdateParams{
		Id:            username,
		EmailVerified: &salt,
	}
	_, err = Repo.UserCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = "Verification successful"
	return resMap, nil
}
