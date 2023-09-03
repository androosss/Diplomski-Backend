// Package backoffice (api -> backoffice) contains code for backoffice specific handling of incoming http requests (version 1).
//
// swagger:operation GET /api-journals ApiJournalCollection
//
// Search api journals.
// ---
// summary: Get api journals by parameters
// operationId: ApiJournalCollection
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
//   - name: createdFrom
//     in: query
//     description: created starting from or after the date.
//     required: false
//     type: string
//   - name: createdBefore
//     in: query
//     description: created before the date.
//     required: false
//     type: string
//   - name: offset
//     in: query
//     description: how many results should be skipped.
//     required: false
//     type: string
//   - name: limit
//     in: query
//     description: max number of results to return. Default is 20.
//     required: false
//     type: string
//   - name: sort
//     in: query
//     description: Allow ascending and descending sorting over multiple fields. Example '+status,-ppAmount'. This returns a list sorted by descending manufacturers and ascending models. Default sort should be -createdAt.
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
//	      $ref: '#/definitions/ApiJournal'
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
package backoffice

import (
	DA "backend/sportos/api/dto"
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"net/http"
)

type ApiJournalsGetHandler struct {
	SearchParams *DR.ApiJournalSearchParams
}

func (r ApiJournalsGetHandler) SupportedMethod() string {
	return http.MethodGet
}

func (r ApiJournalsGetHandler) SupportedSubservers() []DR.SubServer {
	return []DR.SubServer{DR.SUB_BO}
}

func (r *ApiJournalsGetHandler) Init(httpReq *http.Request) DA.Error {
	errorMessages := make([]string, 0)
	var errorMessage string
	r.SearchParams = &DR.ApiJournalSearchParams{}

	r.SearchParams.CreatedAtFrom, errorMessage = DA.ParseDate(DA.GetParameterFromURLQuery(httpReq, "createdFrom"), "createdFrom")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.CreatedAtBefore, errorMessage = DA.ParseDate(DA.GetParameterFromURLQuery(httpReq, "createdBefore"), "createdBefore")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.Offset, errorMessage = DA.ParseInt(DA.GetParameterFromURLQuery(httpReq, "offset"), "offset")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.Limit, errorMessage = DA.ParseInt(DA.GetParameterFromURLQuery(httpReq, "limit"), "limit")
	errorMessages = append(errorMessages, errorMessage)

	r.SearchParams.UserSearchParams = &DR.UserSearchParams{
		//Username: r.username,
	}

	DA.ParseApiJournalSortParams(DA.GetParameterFromURLQuery(httpReq, "sort"), r.SearchParams)

	errorMessages = DA.TrimEmpty(errorMessages)

	if len(errorMessages) > 0 {
		return DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_REQUEST_PARAMS).WithPredefinedPayload(errorMessages)
	}
	return nil
}

func (r *ApiJournalsGetHandler) Validate(ctx context.Context, Repo *crud.Repo) DA.Error {
	return nil
}

func (r *ApiJournalsGetHandler) Process(ctx context.Context, Repo *crud.Repo) (interface{}, DA.Error) {
	res, err := Repo.ApiJournalCrud.Search(ctx, *r.SearchParams, nil)
	if err != nil {
		return res, DA.NewApiError().WithInternalError(err)
	}
	result := make([]DA.ApiJournal, 0)
	for _, aj := range res {
		apiJournal := DA.ApiJournal{}
		apiJournal.InitWithDatabaseStruct(&aj)
		result = append(result, apiJournal)
	}
	resMap := make(map[string]interface{})
	resMap["body"] = result
	cnt, err := Repo.ApiJournalCrud.GetCount(ctx, *r.SearchParams, nil)
	if err != nil {
		return nil, DA.NewApiError().WithInternalError(err)
	}
	resMap["headers"], err = DA.GenerateRangeHeader(cnt, r.SearchParams.PagingSearchParams)
	if err != nil {
		return nil, DA.NewApiError().WithPredefinedError(DA.PRE_ERR_WRONG_RANGE)
	}
	return resMap, nil
}
