package dto

import (
	"backend/sportos/repo/util"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type MatchStatus string

const (
	MS_CREATED  MatchStatus = "CREATED"
	MS_FULL     MatchStatus = "FULL"
	MS_FINISHED MatchStatus = "FINISHED"
)

type Match struct {
	MatchId     string      `json:"matchId,omitempty" column:"match_id"`
	Status      MatchStatus `json:"status,omitempty" column:"status"`
	Players     *string     `json:"players,omitempty" column:"players"`
	PlayerNames *string     `json:"playerNames,omitempty"`
	PlaceId     string      `json:"placeId,omitempty" column:"place_id"`
	Sport       string      `json:"sport,omitempty" column:"sport"`
	StartTime   *time.Time  `json:"startTime,omitempty" column:"start_time"`
	Teams       StrArr      `json:"teams,omitempty" column:"teams"`
	Result      *string     `json:"result" column:"result"`
	EditInfoCUD
}

type StrArr []string

// Value is implementation of data Valuer interface.
func (sa StrArr) Value() (driver.Value, error) {
	return json.Marshal(sa)
}

// Scan is implementation of database/sql scanner interface.
func (sa *StrArr) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &sa)
}

func (s *Match) GetTableName() SportosEntity {
	return "match"
}

func (s *Match) GetId() string {
	return s.MatchId
}

type MatchSearchParams struct {
	Status *string  `json:"status,omitempty"`
	Sports []string `json:"sports,omitempty"`
	EditInfoCUDSearchParams
	MatchSortParams
	PlaceSearchParams *PlaceSearchParams
	PagingSearchParams
	prefix string
}

func (sp *MatchSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "ma"
}

func (sp *MatchSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.MatchSortParams.SetTablePrefix(prefix)
}

func (sp *MatchSearchParams) validate() error {
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

func (sp *MatchSearchParams) joinTables(query *string) {
	if sp.PlaceSearchParams != nil {
		*query += fmt.Sprintf(`inner join place %s on %s.user_id = %s.place_id `, sp.PlaceSearchParams.GetTablePrefix(), sp.PlaceSearchParams.GetTablePrefix(), sp.GetTablePrefix())
		sp.PlaceSearchParams.joinTables(query)
	}
}

func (sp *MatchSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// Place params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.Status != nil && *sp.Status != "" {
		*params = append(*params, *sp.Status)
		*query += fmt.Sprintf(" and %v.status=$%d", tablePrefix, len(*params))
	}
	if len(sp.Sports) > 0 {
		*params = append(*params, pq.Array(sp.Sports))
		*query += fmt.Sprintf(" and %v.sport=any($%d)", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
	// joined params
	if sp.PlaceSearchParams != nil {
		sp.PlaceSearchParams.appendSearchQuery(query, params)
	}
}

func (sp *MatchSearchParams) appendSortQuery(query *string) {
	if !sp.MatchSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.MatchSortParams.OrderBy()
	}
	if sp.PlaceSearchParams != nil {
		sp.PlaceSearchParams.appendSortQuery(query)
	}
}

func (sp *MatchSearchParams) appendGroupByQuery(query *string) {

}

func (sp *MatchSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type MatchSortParams struct {
	Prefix    string
	StartTime *SortColumn `column:"start_time"`
	PlaceId   *SortColumn `column:"place_id"`
	Status    *SortColumn `column:"status"`
	EditInfoCUDSortParams
}

func (sp MatchSortParams) IsEmpty() bool {
	return sp.StartTime == nil && sp.PlaceId == nil && sp.Status == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp MatchSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "ma"
}

func (sp *MatchSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp MatchSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.StartTime != nil {
		sp.StartTime.Prefix = tablePrefix
		sp.StartTime.Column = util.GetTag(sp, "StartTime", column_tag)
		scs = append(scs, *sp.StartTime)
	}
	if sp.PlaceId != nil {
		sp.PlaceId.Prefix = tablePrefix
		sp.PlaceId.Column = util.GetTag(sp, "PlaceId", column_tag)
		scs = append(scs, *sp.PlaceId)
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

func (sp MatchSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type MatchUpdateParams struct {
	Id      string
	Players *string
	Status  *MatchStatus
	Result  *string
	Teams   *StrArr
	EditInfoUDUpdateParams
}

func (up MatchUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update match te set `

	if up.Status != nil {
		*params = append(*params, *up.Status)
		*query += fmt.Sprintf("status = $%d, ", len(*params))
	}

	if up.Players != nil {
		*params = append(*params, up.Players)
		*query += fmt.Sprintf("players = $%d, ", len(*params))
	}

	if up.Result != nil {
		*params = append(*params, up.Result)
		*query += fmt.Sprintf("result = $%d, ", len(*params))
	}

	if up.Teams != nil {
		*params = append(*params, *up.Teams)
		*query += fmt.Sprintf("teams = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where te.match_id = $%d;", len(*params))
}
