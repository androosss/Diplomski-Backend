package dto

import (
	"backend/sportos/repo/util"
	"fmt"
	"strings"
)

type ApiJournal struct {
	ApiJournalId       string  `json:"apiJournal"`
	UserId             *string `json:"playerId,omitempty"`
	Request            *string `json:"request,omitempty"`
	Response           *string `json:"response,omitempty"`
	RequestJson        *string `json:"requestJson,omitempty"`
	ResponseJson       *string `json:"responseJson,omitempty"`
	RequestBodyString  *string
	ResponseBodyString *string
	SourceIP           *string `json:"sourceIP,omitempty"`
	EditInfoCU
}

func (a *ApiJournal) GetTableName() SportosEntity {
	return "api_journal"
}

func (a *ApiJournal) GetId() string {
	return a.ApiJournalId
}

type ApiJournalSearchParams struct {
	ApiJournalId     *string
	SourceIP         *string
	UserSearchParams *UserSearchParams
	ApiJournalSortParams
	EditInfoCSearchParams
	PagingSearchParams
	prefix string
}

func (sp *ApiJournalSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "aj"
}

func (sp *ApiJournalSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.ApiJournalSortParams.SetTablePrefix(prefix)
}

func (sp *ApiJournalSearchParams) validate() error {
	err := sp.EditInfoCSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.PagingSearchParams.validate()
	if err != nil {
		return err
	}
	return nil
}

func (sp *ApiJournalSearchParams) joinTables(query *string) {
	if sp.UserSearchParams != nil {
		*query += fmt.Sprintf(`inner join "user" %s on %s.user_id = %s.user_id `, sp.UserSearchParams.GetTablePrefix(), sp.UserSearchParams.GetTablePrefix(), sp.GetTablePrefix())
		sp.UserSearchParams.joinTables(query)
	}
}

func (sp *ApiJournalSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// api journal params
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	tablePrefix := sp.GetTablePrefix()
	if sp.ApiJournalId != nil && len(*sp.ApiJournalId) != 0 {
		*params = append(*params, *sp.ApiJournalId)
		*query += fmt.Sprintf(" and %v.api_journal_id=$%d", tablePrefix, len(*params))
	}
	if sp.SourceIP != nil && len(*sp.SourceIP) != 0 {
		*params = append(*params, *sp.SourceIP)
		*query += fmt.Sprintf(" and %v.source_ip=$%d", tablePrefix, len(*params))
	}
	if !sp.EditInfoCSearchParams.IsEmpty() {
		sp.EditInfoCSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
	// joined params
	if sp.UserSearchParams != nil {
		sp.UserSearchParams.appendSearchQuery(query, params)
	}
}

func (sp *ApiJournalSearchParams) appendSortQuery(query *string) {
	if !sp.ApiJournalSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.ApiJournalSortParams.OrderBy()
	}
	if sp.UserSearchParams != nil {
		sp.UserSearchParams.appendSortQuery(query)
	}
}

func (sp *ApiJournalSearchParams) appendGroupByQuery(query *string) {

}

func (sp *ApiJournalSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type ApiJournalSortParams struct {
	Prefix       string
	ApiJournalId *SortColumn `column:"api_journal_id"`
	EditInfoCUSortParams
}

func (sp ApiJournalSortParams) IsEmpty() bool {
	return sp.ApiJournalId == nil && sp.EditInfoCSortParams.IsEmpty()
}

func (sp ApiJournalSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "aj"
}

func (sp *ApiJournalSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp ApiJournalSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.ApiJournalId != nil {
		sp.ApiJournalId.Prefix = tablePrefix
		sp.ApiJournalId.Column = util.GetTag(sp, "ApiJournalId", column_tag)
		scs = append(scs, *sp.ApiJournalId)
	}
	if !sp.EditInfoCUSortParams.IsEmpty() {
		sp.EditInfoCUSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUSortParams.SortColumns()...)
	}
	return scs
}

func (sp ApiJournalSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type ApiJournalUpdateParams struct {
	Id           string
	Response     *string
	ResponseJson *string
	UserId       *string
	EditInfoUUpdateParams
}

func (up ApiJournalUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update api_journal aj set `

	if up.Response != nil {
		*params = append(*params, *up.Response)
		*query += fmt.Sprintf("response = $%d, ", len(*params))
	}

	if up.ResponseJson != nil {
		*params = append(*params, *up.ResponseJson)
		*query += fmt.Sprintf("response_json = $%d, ", len(*params))
	}

	if up.UserId != nil {
		*params = append(*params, *up.UserId)
		*query += fmt.Sprintf("user_id = $%d, ", len(*params))
	}

	up.EditInfoUUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where aj.api_journal_id = $%d;", len(*params))
}
