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
//	  description: Conflict error
//	  schema:
//	    $ref: '#/definitions/APIError'
//	500:
//	  description: Internal server error
//	  schema:
//	    $ref: '#/definitions/APIError'
package backoffice

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type AuditsGetHandler struct {
	SearchParams *DR.AuditSearchParams
}

func (r AuditsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r AuditsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_BO}
}

func (r *AuditsGetHandler) Init(httpReq *http.Request) DA.Error {
	errorMessages := make([]string, 0)
	var errorMessage string
	r.SearchParams = &DR.AuditSearchParams{}
	r.SearchParams.Entity = (*DR.SportosEntity)(DA.GetParameterFromURLQuery(httpReq, "entity"))
	r.SearchParams.EntityId = DA.GetParameterFromURLQuery(httpReq, "entityId")

	r.SearchParams.CreatedAtFrom, errorMessage = DA.ParseDate(DA.GetParameterFromURLQuery(httpReq, "createdFrom"), "createdFrom")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.CreatedAtBefore, errorMessage = DA.ParseDate(DA.GetParameterFromURLQuery(httpReq, "createdBefore"), "createdBefore")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.Offset, errorMessage = DA.ParseInt(DA.GetParameterFromURLQuery(httpReq, "offset"), "offset")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.Limit, errorMessage = DA.ParseInt(DA.GetParameterFromURLQuery(httpReq, "limit"), "limit")
	errorMessages = append(errorMessages, errorMessage)

	DA.ParseAuditSortParams(DA.GetParameterFromURLQuery(httpReq, "sort"), r.SearchParams)

	errorMessages = DA.TrimEmpty(errorMessages)

	if len(errorMessages) > 0 {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(errorMessages)
	}
	return nil
}

func (r *AuditsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	if r.SearchParams.EntityId != nil && r.SearchParams.Entity == nil {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_MANDATORY_MISSING).WithMessage("Empty entity field for entityId:'" + *r.SearchParams.EntityId + "'")
	}
	if r.SearchParams.Entity != nil && !r.SearchParams.Entity.IsValid() {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_FORBIDDEN_VALUE).WithMessage("Entity: '" + string(*r.SearchParams.Entity) + "' is not valid")
	} else {
		return nil
	}
}

func (r *AuditsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {

	res, err := Repo.AuditCrud.Search(ctx, *r.SearchParams, nil)
	if err != nil {
		return res, DA.NewApiError().WithInternalError(err)
	}
	result := make([]DA.Audit, 0)
	for _, p := range res {
		apiAudit := DA.Audit{}
		if p.ApiJournalId != nil {
			apiAudit.InitSourceIp(ctx, Repo, *p.ApiJournalId)
		}
		apiAudit.InitWithDatabaseStruct(&p)
		result = append(result, apiAudit)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = result
	cnt, err := Repo.AuditCrud.GetCount(ctx, *r.SearchParams, nil)
	if err != nil {
		return nil, DA.NewApiError().WithInternalError(err)
	}
	resMap["headers"], err = DA.GenerateRangeHeader(cnt, r.SearchParams.PagingSearchParams)
	if err != nil {
		return nil, DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_RANGE)
	}
	return resMap, nil
}
