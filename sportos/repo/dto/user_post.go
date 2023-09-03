package dto

import (
	"fmt"
	"strings"
)

type UserPost struct {
	UserId     string   `json:"userId,omitempty"`
	UserText   string   `json:"userText,omitempty"`
	ImageNames []string `json:"imageNames,omitempty"`
	EditInfoC
}

type UserPostSearchParams struct {
	UserId    *string
	NotUserId *string
	EditInfoCUDSearchParams
	UserPostSortParams
	PagingSearchParams
	prefix string
}

func (sp *UserPostSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "up"
}

func (sp *UserPostSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.UserPostSortParams.SetTablePrefix(prefix)
}

func (sp *UserPostSearchParams) validate() error {
	err := sp.EditInfoCUDSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.PagingSearchParams.validate()
	if err != nil {
		return err
	}
	return nil
}

func (sp *UserPostSearchParams) joinTables(query *string) {
}

func (sp *UserPostSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// UserPost params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.UserId != nil {
		*params = append(*params, *sp.UserId)
		*query += fmt.Sprintf(" and %v.user_id=$%d", tablePrefix, len(*params))
	}
	if sp.NotUserId != nil {
		*params = append(*params, *sp.NotUserId)
		*query += fmt.Sprintf(" and %v.user_id!=$%d", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
}

func (sp *UserPostSearchParams) appendSortQuery(query *string) {
	if !sp.UserPostSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.UserPostSortParams.OrderBy()
	}
}

func (sp *UserPostSearchParams) appendGroupByQuery(query *string) {

}

func (sp *UserPostSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type UserPostSortParams struct {
	Prefix string
	EditInfoCUDSortParams
}

func (sp UserPostSortParams) IsEmpty() bool {
	return sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp UserPostSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "up"
}

func (sp *UserPostSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp UserPostSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp UserPostSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}
