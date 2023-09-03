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
	"encoding/json"
	"net/http"
)

type LogoutPostHandler struct {
	Username string `json:"username"`
}

func (r LogoutPostHandler) SupportedMethod() string {
	return http.MethodPost
}

func (r LogoutPostHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *LogoutPostHandler) Init(httpReq *http.Request) DA.Error {
	decode := json.NewDecoder(httpReq.Body)
	decode.DisallowUnknownFields()
	err := decode.Decode(&r)
	if err == nil {
		return nil
	} else {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(err.Error())
	}
}

func (r *LogoutPostHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	_, err := Repo.UserCrud.GetById(ctx, r.Username, nil)
	if err != nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_ID).WithMessage("Username doesn't exist")
	}
	return nil
}

func (r *LogoutPostHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	up := DR.UserUpdateParams{
		Id:                r.Username,
		Token:             nil,
		TokenValidUntil:   nil,
		TokenRefreshUntil: nil,
	}
	_, err := Repo.UserCrud.Update(ctx, up, nil, nil)
	if err != nil {
		return nil, DA.InternalServerError(err)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = struct{}{}
	return resMap, nil
}
