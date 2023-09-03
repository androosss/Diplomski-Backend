package dto

import (
	"backend/sportos/repo/util"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type PracticeStatus string

const (
	PS_CREATED  PracticeStatus = "CREATED"
	PS_ACCEPTED PracticeStatus = "ACCEPTED"
	PS_DENIED   PracticeStatus = "DENIED"
)

type Practice struct {
	PracticeId string         `json:"practiceId,omitempty" column:"practice_id"`
	Status     PracticeStatus `json:"status,omitempty" column:"status"`
	PlayerId   string         `json:"playerId,omitempty" column:"player_id"`
	CoachId    string         `json:"coachId,omitempty" column:"coach_id"`
	Sport      string         `json:"sport,omitempty" column:"sport"`
	StartTime  *time.Time     `json:"startTime,omitempty" column:"start_time"`
	EditInfoCUD
}

func (s *Practice) GetTableName() SportosEntity {
	return "practice"
}

func (s *Practice) GetId() string {
	return s.PracticeId
}

type PracticeSearchParams struct {
	PlayerId *string  `json:"playerId,omitempty"`
	CoachId  *string  `json:"coachId,omitempty"`
	Status   *string  `json:"status,omitempty"`
	Sports   []string `json:"sports,omitempty"`
	EditInfoCUDSearchParams
	PracticeSortParams
	PagingSearchParams
	prefix string
}

func (sp *PracticeSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "pr"
}

func (sp *PracticeSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.PracticeSortParams.SetTablePrefix(prefix)
}

func (sp *PracticeSearchParams) validate() error {
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

func (sp *PracticeSearchParams) joinTables(query *string) {
}

func (sp *PracticeSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// Place params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.Status != nil && *sp.Status != "" {
		*params = append(*params, *sp.Status)
		*query += fmt.Sprintf(" and %v.status=$%d", tablePrefix, len(*params))
	}
	if sp.PlayerId != nil {
		*params = append(*params, *sp.PlayerId)
		*query += fmt.Sprintf(" and %v.player_id=$%d", tablePrefix, len(*params))
	}
	if sp.CoachId != nil {
		*params = append(*params, *sp.CoachId)
		*query += fmt.Sprintf(" and %v.coach_id=$%d", tablePrefix, len(*params))
	}
	if len(sp.Sports) > 0 {
		*params = append(*params, pq.Array(sp.Sports))
		*query += fmt.Sprintf(" and %v.sport=any($%d)", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
}

func (sp *PracticeSearchParams) appendSortQuery(query *string) {
	if !sp.PracticeSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.PracticeSortParams.OrderBy()
	}
}

func (sp *PracticeSearchParams) appendGroupByQuery(query *string) {

}

func (sp *PracticeSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type PracticeSortParams struct {
	Prefix    string
	StartTime *SortColumn `column:"start_time"`
	Status    *SortColumn `column:"status"`
	EditInfoCUDSortParams
}

func (sp PracticeSortParams) IsEmpty() bool {
	return sp.StartTime == nil && sp.Status == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp PracticeSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "pr"
}

func (sp *PracticeSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp PracticeSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.StartTime != nil {
		sp.StartTime.Prefix = tablePrefix
		sp.StartTime.Column = util.GetTag(sp, "StartTime", column_tag)
		scs = append(scs, *sp.StartTime)
	}
	if sp.Status != nil {
		sp.Status.Prefix = tablePrefix
		sp.Status.Column = util.GetTag(sp, "Status", column_tag)
		scs = append(scs, *sp.Status)
	}
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp PracticeSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type PracticeUpdateParams struct {
	Id     string
	Status *PracticeStatus
	EditInfoUDUpdateParams
}

func (up PracticeUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update practice te set `

	if up.Status != nil {
		*params = append(*params, *up.Status)
		*query += fmt.Sprintf("status = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where te.practice_id = $%d;", len(*params))
}
