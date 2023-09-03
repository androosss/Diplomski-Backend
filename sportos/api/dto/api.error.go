package dto

import (
	L "backend/internal/logging"
	"encoding/json"
	"net/http"
	"runtime/debug"
)

// Error is an interface for Api errors
type Error interface {
	error
	IsEmpty() bool
	WithInternalError(err error) Error
	WithPredefinedError(err PredefinedError) Error
	WithPredefinedPayload(payload interface{}) Error
	WithMessage(message string) Error
	GetInternalError() error
	GetPredefinedError() PredefinedError
	GetHTTPCode() int
	GetMessage() string
	GetPredefinedPayload() json.RawMessage
	HasInternalError() bool
	HasPredefinedError() bool
}

// ApiError error structure
type ApiError struct {
	// Predefined Api error code
	Predefined PredefinedError `json:"errorCode,omitempty"`
	// Predefined payload raw data
	Payload json.RawMessage `json:"payload,omitempty"`
	// internal error data
	Internal error `json:"internal,omitempty"`
	// error message with additional details
	Message string `json:"errorMessage,omitempty"`
	// Http code
	Code int `json:"code,omitempty"`
}

// Implementations of error interface

// NewApiError creates new Api error
func NewApiError() Error {
	return new(ApiError)
}

func InternalServerError(err error) Error {
	L.L.Error("Internal server error", L.Error(err))
	debug.PrintStack()
	return &ApiError{Code: 500}
}

func ErrorNotFound() Error {
	return &ApiError{Code: 404}
}

func ErrorMethodNotAllowed() Error {
	return &ApiError{Code: 405}
}

func ErrorBadRequest() Error {
	return &ApiError{Code: 400}
}

func ErrorUnauthorized() Error {
	return &ApiError{Code: 401}
}

func ErrorForbidden() Error {
	return &ApiError{Code: 403}
}

func (ae *ApiError) Error() string {
	if ae.Predefined != "" {
		return string(ae.Predefined)
	}
	if ae.Internal != nil {
		return ae.Error()
	}
	if ae.Code != 0 {
		return http.StatusText(ae.Code)
	}
	L.L.Fatal("Error ins't initialized!")
	return ""
}

// ApiError add internal error
func (ae *ApiError) WithInternalError(err error) Error {
	L.L.Error("PredefinedError WithInternalError", L.Any("error", err))
	debug.PrintStack()
	if ae.Predefined != "" {
		L.L.Fatal("Predefined error already added. Internal error can not be added too.")
	}
	if err != nil {
		ae.Internal = err
	} else {
		L.L.Fatal("Throwing nil TRI Pay internal error is not allowed!")
	}
	return ae
}

// ApiError add predefined Api error
func (ae *ApiError) WithPredefinedError(err PredefinedError) Error {
	if err != "" {
		ae.Predefined = err
	}
	return ae
}

// ApiError add predefined message
func (ae *ApiError) WithMessage(message string) Error {
	if message != "" {
		ae.Message = message
	}
	return ae
}

// ApiError adds a payload raw data to an error
func (ae *ApiError) WithPredefinedPayload(payload interface{}) Error {
	if ae.Predefined == "" {
		L.L.Fatal("Set Predefined Error first when adding payload")
	}
	if payload != nil {
		p, err := json.Marshal(payload)
		if err != nil {
			L.L.Error("ApiError.WithPredefinePayload", L.Any("payload", payload), L.Error(err))
			return ae
		}
		ae.Payload = p
	}
	return ae
}

// ApiError get internal error
func (ae *ApiError) GetInternalError() error {
	return ae.Internal
}

// ApiError get predefined error
func (ae *ApiError) GetPredefinedError() PredefinedError {
	return ae.Predefined
}

// ApiError get HTTP code
func (ae *ApiError) GetHTTPCode() int {
	if ae.Code != 0 {
		return ae.Code
	}
	if ae.Predefined != "" {
		return ApiErrorHTTPCodesMap[ae.Predefined]
	}
	if ae.Internal != nil {
		return http.StatusInternalServerError
	}
	return 400
}

// ApiError get message
func (ae *ApiError) GetMessage() string {
	return ae.Message
}

// ApiError get payload
func (ae *ApiError) GetPredefinedPayload() json.RawMessage {
	return ae.Payload
}

// Api error check if it has body to print
func (ae *ApiError) IsEmpty() bool {
	return ae.Message == "" && ae.Payload == nil && ae.Predefined == ""
}

// ApiError check if it has internal error
func (ae *ApiError) HasInternalError() bool {
	return ae.Internal != nil
}

// ApiError check if it has predefined error
func (ae *ApiError) HasPredefinedError() bool {
	return ae.Predefined != ""
}

// [swagger]

// PredefinedError
//
// Predefined errors. Possible error messages:
//   - 'mandatory_field_missing' - missing mandatory field in payment instrument values
//   - 'forbidden_id' - non existing entity id sent from frontend
//   - 'forbidden_value' - non existing enum sent from frontend
//   - 'unique_constraint' - entity may not be created due to its unique constraints
//   - 'wrong_request_parametars' - wrong parameters received from body or URI
//   - 'wrong_range_parametars' - wrong range parameters received from body or URI
//
// swagger:model PredefinedError
type PredefinedError string

// Errors returned by Api package
const (
	// missing mandatory field in payment instrument values
	PRE_ERR_MANDATORY_MISSING PredefinedError = "mandatory_field_missing"
	// non existing object id sent from frontend
	PRE_ERR_FORBIDDEN_ID PredefinedError = "forbidden_id"
	// non existing enum sent from frontend
	PRE_ERR_FORBIDDEN_VALUE PredefinedError = "forbidden_value"
	// entity may not be created due to its unique constraints
	PRE_ERR_UNIQUE_CONSTRAINT PredefinedError = "unique_constraint"
	// wrong parameters received from body or URI
	PRE_ERR_WRONG_REQUEST_PARAMS PredefinedError = "wrong_request_parametars"
	// wrong range parameters received
	PRE_ERR_WRONG_RANGE PredefinedError = "wrong_range_parameters"
	// user didn't verify email
	PRE_ERR_MAIL_NOT_VERIFIED PredefinedError = "mail_not_verified"
	// parameters have bad format
	PRE_ERR_BAD_FORMAT PredefinedError = "bad_format"
)

// default is 404 (Not Found) if not set
var ApiErrorHTTPCodesMap = map[PredefinedError]int{
	PRE_ERR_WRONG_REQUEST_PARAMS: http.StatusBadRequest,
	PRE_ERR_UNIQUE_CONSTRAINT:    http.StatusBadRequest,
	PRE_ERR_WRONG_RANGE:          http.StatusBadRequest,
	PRE_ERR_FORBIDDEN_VALUE:      http.StatusBadRequest,
	PRE_ERR_MANDATORY_MISSING:    http.StatusBadRequest,
	PRE_ERR_FORBIDDEN_ID:         http.StatusBadRequest,
	PRE_ERR_MAIL_NOT_VERIFIED:    http.StatusMethodNotAllowed,
	PRE_ERR_BAD_FORMAT:           http.StatusBadRequest,
}
