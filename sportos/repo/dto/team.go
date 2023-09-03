package dto

import (
	"backend/sportos/repo/util"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type TeamStatus string

const (
	TS_CREATED    TeamStatus = "ACTIVE"
	TS_FULL       TeamStatus = "FULL"
	TS_REGISTERED TeamStatus = "INACTIVE"
)

type Team struct {
	TeamId  string     `json:"teamId" column:"team_id"`
	Name    string     `json:"name" column:"name"`
	Sport   string     `json:"sport" column:"sport"`
	Status  TeamStatus `json:"status" column:"status"`
	Players string     `json:"players" column:"players"`
	EditInfoCUD
}

func (s *Team) GetTableName() SportosEntity {
	return "team"
}

func (s *Team) GetId() string {
	return s.TeamId
}

type TeamSearchParams struct {
	Name          *string  `json:"name,omitempty"`
	Sports        []string `json:"sport,omitempty"`
	UserNotInTeam *string  `json:"userNotInTeam,omitempty"`
	Owner         *string  `json:"owner,omitempty"`
	Status        *string  `json:"status,omitempty"`
	EditInfoCUDSearchParams
	TeamSortParams
	PagingSearchParams
	prefix string
}

func (sp *TeamSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "te"
}

func (sp *TeamSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.TeamSortParams.SetTablePrefix(prefix)
}

func (sp *TeamSearchParams) validate() error {
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

func (sp *TeamSearchParams) joinTables(query *string) {
}

func (sp *TeamSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// Place params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.Name != nil && *sp.Name != "" {
		*params = append(*params, *sp.Name)
		*query += fmt.Sprintf(" and %v.name=$%d", tablePrefix, len(*params))
	}
	if sp.Status != nil && *sp.Status != "" {
		*params = append(*params, *sp.Status)
		*query += fmt.Sprintf(" and %v.status=$%d", tablePrefix, len(*params))
	}
	if len(sp.Sports) != 0 {
		*params = append(*params, pq.Array(sp.Sports))
		*query += fmt.Sprintf(" and %v.sport=any($%d)", tablePrefix, len(*params))
	}
	if sp.UserNotInTeam != nil {
		*params = append(*params, *sp.UserNotInTeam)
		*query += fmt.Sprintf(" and %v.players not like '%%'||$%d||'%%'", tablePrefix, len(*params))
	}
	if sp.Owner != nil {
		*params = append(*params, *sp.Owner)
		*query += fmt.Sprintf(" and %v.players like $%d||'%%'", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
}

func (sp *TeamSearchParams) appendSortQuery(query *string) {
	if !sp.TeamSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.TeamSortParams.OrderBy()
	}
}

func (sp *TeamSearchParams) appendGroupByQuery(query *string) {

}

func (sp *TeamSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type TeamSortParams struct {
	Prefix string
	Name   *SortColumn `column:"name"`
	EditInfoCUDSortParams
}

func (sp TeamSortParams) IsEmpty() bool {
	return sp.Name == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp TeamSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "te"
}

func (sp *TeamSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp TeamSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.Name != nil {
		sp.Name.Prefix = tablePrefix
		sp.Name.Column = util.GetTag(sp, "Name", column_tag)
		scs = append(scs, *sp.Name)
	}
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp TeamSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type TeamUpdateParams struct {
	Id      string
	Players *string
	Status  *TeamStatus
	EditInfoUDUpdateParams
}

func (up TeamUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update team te set `

	if up.Status != nil {
		*params = append(*params, *up.Status)
		*query += fmt.Sprintf("status = $%d, ", len(*params))
	}

	if up.Players != nil {
		*params = append(*params, up.Players)
		*query += fmt.Sprintf("players = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where te.team_id = $%d;", len(*params))
}
