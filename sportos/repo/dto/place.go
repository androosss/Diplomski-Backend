package dto

import (
	"backend/sportos/repo/util"
	"fmt"
	"strings"
)

type Place struct {
	Username string   `json:"username" column:"user_id"`
	Name     string   `json:"name" column:"name"`
	City     string   `json:"city" column:"city"`
	Sport    string   `json:"sport" column:"sport"`
	Booking  *Booking `json:"booking" column:"booking"`
	Reviews  *Reviews `json:"reviews" column:"reviews"`
	EditInfoCUD
}

func (s *Place) GetTableName() SportosEntity {
	return "place"
}

func (s *Place) GetId() string {
	return s.Username
}

type PlaceSearchParams struct {
	Username *string `json:"username,omitempty"`
	Name     *string `json:"name,omitempty"`
	City     *string `json:"city,omitempty"`
	Sport    *string `json:"sport" column:"sport"`
	EditInfoCUDSearchParams
	PlaceSortParams
	UserSearchParams *UserSearchParams
	PagingSearchParams
	prefix string
}

func (sp *PlaceSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "pla"
}

func (sp *PlaceSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.PlaceSortParams.SetTablePrefix(prefix)
}

func (sp *PlaceSearchParams) validate() error {
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

func (sp *PlaceSearchParams) joinTables(query *string) {
	if sp.UserSearchParams != nil {
		*query += fmt.Sprintf(`inner join "user" %s on %s.user_id = %s.user_id `, sp.UserSearchParams.GetTablePrefix(), sp.UserSearchParams.GetTablePrefix(), sp.GetTablePrefix())
		sp.UserSearchParams.joinTables(query)
	}
}

func (sp *PlaceSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// Place params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.Username != nil && *sp.Username != "" {
		*params = append(*params, *sp.Username)
		*query += fmt.Sprintf(" and %v.user_id=$%d", tablePrefix, len(*params))
	}
	if sp.Name != nil && *sp.Name != "" {
		*params = append(*params, *sp.Name)
		*query += fmt.Sprintf(" and %v.name=$%d", tablePrefix, len(*params))
	}
	if sp.City != nil && *sp.City != "" {
		*params = append(*params, *sp.City)
		*query += fmt.Sprintf(" and %v.city=$%d", tablePrefix, len(*params))
	}
	if sp.Sport != nil && *sp.Sport != "" {
		*params = append(*params, *sp.Sport)
		*query += fmt.Sprintf(" and %v.sport=$%d", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
	// joined params
	if sp.UserSearchParams != nil {
		sp.UserSearchParams.appendSearchQuery(query, params)
	}
}

func (sp *PlaceSearchParams) appendSortQuery(query *string) {
	if !sp.PlaceSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.PlaceSortParams.OrderBy()
	}
	if sp.UserSearchParams != nil {
		sp.UserSearchParams.appendSortQuery(query)
	}
}

func (sp *PlaceSearchParams) appendGroupByQuery(query *string) {

}

func (sp *PlaceSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type PlaceSortParams struct {
	Prefix   string
	Username *SortColumn `column:"username"`
	Name     *SortColumn `column:"name"`
	City     *SortColumn `column:"city"`
	EditInfoCUDSortParams
}

func (sp PlaceSortParams) IsEmpty() bool {
	return sp.Username == nil && sp.Name == nil && sp.City == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp PlaceSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "pla"
}

func (sp *PlaceSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp PlaceSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.Username != nil {
		sp.Username.Prefix = tablePrefix
		sp.Username.Column = util.GetTag(sp, "Username", column_tag)
		scs = append(scs, *sp.Username)
	}
	if sp.Name != nil {
		sp.Name.Prefix = tablePrefix
		sp.Name.Column = util.GetTag(sp, "Name", column_tag)
		scs = append(scs, *sp.Name)
	}
	if sp.City != nil {
		sp.City.Prefix = tablePrefix
		sp.City.Column = util.GetTag(sp, "City", column_tag)
		scs = append(scs, *sp.City)
	}
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp PlaceSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type PlaceUpdateParams struct {
	Id      string
	Name    *string
	City    *string
	Booking *Booking
	Reviews *Reviews
	EditInfoUDUpdateParams
}

func (up PlaceUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update place pla set `

	if up.Name != nil {
		*params = append(*params, *up.Name)
		*query += fmt.Sprintf("name = $%d, ", len(*params))
	}

	if up.City != nil {
		*params = append(*params, up.City)
		*query += fmt.Sprintf("city = $%d, ", len(*params))
	}

	if up.Booking != nil {
		*params = append(*params, up.Booking)
		*query += fmt.Sprintf("booking = $%d, ", len(*params))
	}

	if up.Reviews != nil {
		*params = append(*params, up.Reviews)
		*query += fmt.Sprintf("reviews = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where pla.user_id = $%d;", len(*params))
}
