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

type Standing struct {
	TeamName string `json:"teamName,omitempty"`
	Points   *int   `json:"points,omitempty"`
	Ranking  *int   `json:"ranking,omitempty"`
}

type Pairing struct {
	TeamOne string `json:"teamOne,omitempty"`
	TeamTwo string `json:"teamTwo,omitempty"`
	Score   string `json:"score"`
}

type Round struct {
	Pairing []Pairing `json:"pairing,omitempty"`
}

type Tournament struct {
	Standings []Standing `json:"standings,omitempty"`
	Rounds    []Round    `json:"rounds,omitempty"`
}

type TeamRef struct {
	Name   string `json:"name,omitempty"`
	TeamId string `json:"teamId,omitempty"`
}

type EventStatus string

const (
	ES_CREATED   EventStatus = "CREATED"
	ES_ACTIVE    EventStatus = "ACTIVE"
	ES_FINISHED  EventStatus = "FINISHED"
	ES_CANCELLED EventStatus = "CANCELLED"
)

type Event struct {
	EventId    string      `json:"eventId,omitempty" column:"event_id"`
	Name       string      `json:"name,omitempty" column:"name"`
	Owner      string      `json:"owner,omitempty" column:"owner_id"`
	Sport      string      `json:"sport,omitempty" column:"sport"`
	Status     EventStatus `json:"eventStatus,omitempty" column:"status"`
	Time       *time.Time  `json:"time,omitempty" column:"time"`
	Teams      Teams       `json:"teams,omitempty" column:"teams"`
	Tournament *Tournament `json:"tournament,omitempty" column:"tournament"`
	EditInfoCUD
}

func (e *Event) GetId() string {
	return e.EventId
}

func (e *Event) GetTableName() SportosEntity {
	return "event"
}

// Value is implementation of data Valuer interface.
func (t Tournament) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan is implementation of database/sql scanner interface.
func (t *Tournament) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &t)
}

type Teams []TeamRef

// Value is implementation of data Valuer interface.
func (t Teams) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan is implementation of database/sql scanner interface.
func (t *Teams) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &t)
}

type EventSearchParams struct {
	Name   *string  `json:"name,omitempty"`
	Owner  *string  `json:"owner,omitempty"`
	Sports []string `json:"sports,omitempty"`
	Status *string  `json:"status,omitempty"`
	EditInfoCUDSearchParams
	EventSortParams
	PagingSearchParams
	prefix string
}

func (sp *EventSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "ev"
}

func (sp *EventSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.EventSortParams.SetTablePrefix(prefix)
}

func (sp *EventSearchParams) validate() error {
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

func (sp *EventSearchParams) joinTables(query *string) {
}

func (sp *EventSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// Place params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.Name != nil && *sp.Name != "" {
		*params = append(*params, *sp.Name)
		*query += fmt.Sprintf(" and %v.name=$%d", tablePrefix, len(*params))
	}
	if sp.Owner != nil && *sp.Owner != "" {
		*params = append(*params, *sp.Owner)
		*query += fmt.Sprintf(" and %v.owner_id=$%d", tablePrefix, len(*params))
	}
	if len(sp.Sports) != 0 {
		*params = append(*params, pq.Array(sp.Sports))
		*query += fmt.Sprintf(" and %v.sport=any($%d)", tablePrefix, len(*params))
	}
	if sp.Status != nil {
		*params = append(*params, *sp.Status)
		*query += fmt.Sprintf(" and %v.status=$%d", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
}

func (sp *EventSearchParams) appendSortQuery(query *string) {
	if !sp.EventSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.EventSortParams.OrderBy()
	}
}

func (sp *EventSearchParams) appendGroupByQuery(query *string) {

}

func (sp *EventSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type EventSortParams struct {
	Prefix string
	Owner  *SortColumn `column:"owner_id"`
	Sport  *SortColumn `column:"sport"`
	EditInfoCUDSortParams
}

func (sp EventSortParams) IsEmpty() bool {
	return sp.Owner == nil && sp.Sport == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp EventSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "ev"
}

func (sp *EventSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp EventSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.Owner != nil {
		sp.Owner.Prefix = tablePrefix
		sp.Owner.Column = util.GetTag(sp, "Owner", column_tag)
		scs = append(scs, *sp.Owner)
	}
	if sp.Sport != nil {
		sp.Sport.Prefix = tablePrefix
		sp.Sport.Column = util.GetTag(sp, "Sport", column_tag)
		scs = append(scs, *sp.Sport)
	}
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp EventSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type EventUpdateParams struct {
	Id         string
	Teams      *Teams
	Status     *EventStatus
	Tournament *Tournament
	EditInfoUDUpdateParams
}

func (up EventUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update event ev set `

	if up.Teams != nil {
		*params = append(*params, *up.Teams)
		*query += fmt.Sprintf("teams = $%d, ", len(*params))
	}

	if up.Tournament != nil {
		*params = append(*params, up.Tournament)
		*query += fmt.Sprintf("tournament = $%d, ", len(*params))
	}

	if up.Status != nil {
		*params = append(*params, up.Status)
		*query += fmt.Sprintf("status = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where ev.event_id = $%d;", len(*params))
}
