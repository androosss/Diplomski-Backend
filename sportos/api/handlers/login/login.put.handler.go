package login

// swagger:operation POST /partner-payment-provider-profiles PartnerPaymentProviderProfileCreate
//
// Create a new payment provider merchant profile for partner
// ---
// summary: Create payment provider profile for partner
// operationId: PartnerPaymentProviderProfileCreate
// tags:
// - Backoffice API
// produces:
// - application/json
// parameters:
//   - name: X-Request-ID
//     in: header
//     description: Unique request ID
//     required: true
//     type: string
//   - name: User-Id
//     in: header
//     description: Id of backoffice user
//     required: true
//     type: string
//   - in: body
//     name: partnerPaymentProviderProfile
//     description: profile that should be created for partner
//     required: true
//     schema:
//     $ref: '#/definitions/PartnerPaymentProviderProfile'
//
// responses:
//
//	200:
//	  description: results
//	  schema:
//	    $ref: '#/definitions/PartnerPaymentProviderProfile'
//	400:
//	  description: Invalid request supplied
//	  schema:
//	    $ref: '#/definitions/APIError'
//	403:
//	  description: Request is forbidden
//	  schema:
//	    $ref: '#/definitions/APIError'
//	404:
//	  description: Predefined error occured
//	  schema:
//	    $ref: '#/definitions/APIError'
//	405:
//	  description: Method not allowed error
//	  schema:
//	    $ref: '#/definitions/APIError'
//	409:
//	  description: Conflict error, payment provider error
//	  schema:
//	    $ref: '#/definitions/APIError'
//	500:
//	  description: Internal server error
//	  schema:
//	    $ref: '#/definitions/APIError'

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

type LoginPutRequest struct {
	Username *string `json:"username"`
	Token    *string `json:"accessToken"`
}

type LoginPutHandler struct {
	LoginPutRequest
}

type LoginPutResponse struct {
	Username    string `json:"username,omitempty"`
	Type        string `json:"type,omitempty"`
	City        string `json:"city,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
}

func (r LoginPutHandler) SupportedMethod() string {
	return http.MethodPut
}

func (r LoginPutHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *LoginPutHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r.LoginPutRequest)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *LoginPutHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.Username == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Username is mandatory")
	}
	if r.Token == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Token is mandatory")
	}
	return nil
}

func (r *LoginPutHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sha256Hasher := sha256.New()
	user, err := Repo.UserCrud.GetById(ctx, *r.Username, nil)
	if err != nil {
		return nil, DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Username doesn't exist")
	}
	token := *r.Username
	validUntil := time.Now().Add(20 * time.Minute)
	if validUntil.After(*user.TokenRefreshUntil) {
		return nil, DA.ErrorUnauthorized().WithMessage("token can't  be refreshed, it expired")
	}
	token += validUntil.Format(time.RFC3339Nano)
	data, _ := json.Marshal(token)
	sha256Hasher.Reset()
	sha256Hasher.Write(data)
	token = base64.URLEncoding.EncodeToString(sha256Hasher.Sum(nil))
	up := DR.UserUpdateParams{
		Id:              *r.Username,
		Token:           &token,
		TokenValidUntil: &validUntil,
	}
	_, err = Repo.UserCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	city := ""
	switch user.UserType {
	case DR.UT_PLACE:
		pla, _ := Repo.PlaceCrud.GetById(ctx, *r.Username, nil)
		city = pla.City
	case DR.UT_PLAYER:
		pl, _ := Repo.PlayerCrud.GetById(ctx, *r.Username, nil)
		city = pl.City
	case DR.UT_COACH:
		co, _ := Repo.CoachCrud.GetById(ctx, *r.Username, nil)
		city = co.City
	}
	resMap := make(map[string]interface{})
	resMap["body"] = LoginPutResponse{AccessToken: token, Type: string(user.UserType), City: city, Username: user.Username}
	return resMap, nil
}
