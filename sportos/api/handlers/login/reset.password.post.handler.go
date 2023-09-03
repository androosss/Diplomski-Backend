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

package login

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

type ResetPasswordPostHandler struct {
	ResetPasswordPostRequest
	Username string
}

type ResetPasswordPostRequest struct {
	Password   string `json:"password,omitempty"`
	ResetToken string `json:"resetToken,omitempty"`
}

func (r ResetPasswordPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r ResetPasswordPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *ResetPasswordPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *ResetPasswordPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	token, err := base64.URLEncoding.DecodeString(r.ResetToken)
	if err != nil {
		return DA.ErrorBadRequest()
	}
	decoded := strings.Split(string(token), "___")
	if len(decoded) != 2 {
		return DA.ErrorBadRequest()
	}
	username, password := decoded[0], decoded[1]
	user, err := Repo.UserCrud.GetById(ctx, username, nil)
	if err != nil {
		return DA.ErrorBadRequest()
	}
	if user.PasswordHash != password {
		return DA.ErrorBadRequest()
	}
	if !verifyPassword(r.Password) {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Password isn't strong enough")
	}
	r.Username = username
	return nil
}

func (r *ResetPasswordPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	sha256Hasher := sha256.New()
	data, _ := json.Marshal(r.Password)
	sha256Hasher.Write(data)
	dataHash := base64.URLEncoding.EncodeToString(sha256Hasher.Sum(nil))
	up := DR.UserUpdateParams{
		Id:           r.Username,
		PasswordHash: &dataHash,
	}
	_, err := Repo.UserCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = struct{}{}
	return resMap, nil
}
