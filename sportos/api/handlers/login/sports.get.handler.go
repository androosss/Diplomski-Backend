// swagger:operation GET /audits AuditCollection
//
// Search audits.
// ---
// summary: Get audits by parameters
// operationId: AuditCollection
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
//   - name: entity
//     in: query
//     description: name of database entity. Value is from enum SportosEntity
//     required: false
//     type: string
//   - name: entityId
//     in: query
//     description: id of database entity. Entity is mandatory if entityId exists.
//     required: false
//     type: string
//
// responses:
//
//	200:
//	  description: results
//	  schema:
//	    type: array
//	    items:
//	      $ref: '#/definitions/Audit'
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
	"net/http"
)

type SportsGetHandler struct {
}

func (r SportsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r SportsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_LO}
}

func (r *SportsGetHandler) Init(httpReq *http.Request) DA.Error {
	return nil
}

func (r *SportsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *SportsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	result := DR.Sports
	resMap := make(map[string]interface{})
	resMap["body"] = result
	cnt := len(result)
	resMap["headers"], _ = DA.GenerateRangeHeader(cnt, DR.PagingSearchParams{})
	return resMap, nil
}
