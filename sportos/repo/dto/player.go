package dto

import (
	"backend/sportos/repo/util"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Player struct {
	Username    string      `json:"username" column:"user_id"`
	Name        string      `json:"name" column:"name"`
	City        string      `json:"city" column:"city"`
	Preferences *Prefernces `json:"preferences" column:"preferences"`
	Statistics  StatMap     `json:"statistics" column:"statistics"`
	EditInfoCUD
}

func (s *Player) GetTableName() SportosEntity {
	return "player"
}

func (s *Player) GetId() string {
	return s.Username
}

type PlayerSearchParams struct {
	Username *string `json:"username,omitempty"`
	Name     *string `json:"name,omitempty"`
	City     *string `json:"city,omitempty"`
	EditInfoCUDSearchParams
	PlayerSortParams
	UserSearchParams *UserSearchParams
	PagingSearchParams
	prefix string
}

func (sp *PlayerSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "pl"
}

func (sp *PlayerSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.PlayerSortParams.SetTablePrefix(prefix)
}

func (sp *PlayerSearchParams) validate() error {
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

func (sp *PlayerSearchParams) joinTables(query *string) {
	if sp.UserSearchParams != nil {
		*query += fmt.Sprintf(`inner join "user" %s on %s.user_id = %s.user_id `, sp.UserSearchParams.GetTablePrefix(), sp.UserSearchParams.GetTablePrefix(), sp.GetTablePrefix())
		sp.UserSearchParams.joinTables(query)
	}
}

func (sp *PlayerSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// player params
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
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
	// joined params
	if sp.UserSearchParams != nil {
		sp.UserSearchParams.appendSearchQuery(query, params)
	}
}

func (sp *PlayerSearchParams) appendSortQuery(query *string) {
	if !sp.PlayerSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.PlayerSortParams.OrderBy()
	}
	if sp.UserSearchParams != nil {
		sp.UserSearchParams.appendSortQuery(query)
	}
}

func (sp *PlayerSearchParams) appendGroupByQuery(query *string) {

}

func (sp *PlayerSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type PlayerSortParams struct {
	Prefix   string
	Username *SortColumn `column:"username"`
	Name     *SortColumn `column:"name"`
	City     *SortColumn `column:"city"`
	EditInfoCUDSortParams
}

func (sp PlayerSortParams) IsEmpty() bool {
	return sp.Username == nil && sp.Name == nil && sp.City == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp PlayerSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "pl"
}

func (sp *PlayerSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp PlayerSortParams) SortColumns() SortColumns {
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

func (sp PlayerSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type PlayerUpdateParams struct {
	Id          string
	Name        *string
	City        *string
	Preferences *Prefernces
	Statistics  *StatMap
	EditInfoUDUpdateParams
}

func (up PlayerUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update player pl set `

	if up.Name != nil {
		*params = append(*params, *up.Name)
		*query += fmt.Sprintf("name = $%d, ", len(*params))
	}

	if up.City != nil {
		*params = append(*params, up.City)
		*query += fmt.Sprintf("city = $%d, ", len(*params))
	}

	if up.Preferences != nil {
		*params = append(*params, up.Preferences)
		*query += fmt.Sprintf("preferences = $%d, ", len(*params))
	}

	if up.Statistics != nil {
		*params = append(*params, *up.Statistics)
		*query += fmt.Sprintf("statistics = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where pl.user_id = $%d;", len(*params))
}

type PlayerLevel string

const (
	LEVEL_PRO PlayerLevel = "Proffesional"
	LEVEL_AMA PlayerLevel = "Amateur"
)

type Prefernce struct {
	Sport    string      `json:"sport,omitempty"`
	Position string      `json:"position,omitempty"`
	Level    PlayerLevel `json:"level,omitempty"`
}

type Prefernces []Prefernce

// Value is implementation of data Valuer interface.
func (p Prefernces) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan is implementation of database/sql scanner interface.
func (p *Prefernces) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

type Statistic struct {
	MyTeam  []string  `json:"myTeam,omitempty"`
	OppTeam []string  `json:"oppTeam,omitempty"`
	Date    time.Time `json:"date,omitempty"`
	Score   string    `json:"score,omitempty"`
}

type TournamentFinish struct {
	MyTeam     []string `json:"myTeam,omitempty"`
	Ranking    int      `json:"ranking"`
	Tournament string   `json:"tournament"`
}

type Statistics struct {
	Matches     []Statistic        `json:"matches,omitempty"`
	WinRatio    decimal.Decimal    `json:"winRatio,omitempty"`
	Tournaments []TournamentFinish `json:"tournaments,omitempty"`
}

type StatMap map[string]Statistics

// Value is implementation of data Valuer interface.
func (s StatMap) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan is implementation of database/sql scanner interface.
func (s *StatMap) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

// Value is implementation of data Valuer interface.
func (s Statistics) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan is implementation of database/sql scanner interface.
func (s *Statistics) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}
