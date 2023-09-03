package dto

import (
	"backend/sportos/repo/util"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	system_user     = "system"
	SCHEDULE_RUNNER = "schedule_runner"
	column_tag      = "column"
)

// This interface represents one database table implementation
type CommonEntity interface {
	GetId() string
	GetTableName() SportosEntity
}

// This interface represents implementations for constructing select queries for database
type SearchParams interface {
	GetTablePrefix() string
	SetTablePrefix(string)
	validate() error
	joinTables(query *string)
	appendSearchQuery(query *string, params *[]interface{})
	appendSortQuery(query *string)
	appendGroupByQuery(query *string)
	appendPagingQuery(query *string, params *[]interface{})
}

// This interface represents implementations for constructing update queries for database
type UpdateParams interface {
	appendUpdateQuery(query *string, params *[]interface{})
}

func (ei EditInfoCUD) IsDeleted() bool {
	return ei.DeletedAt == nil
}

func (ei EditInfoCD) IsDeleted() bool {
	return ei.DeletedAt == nil
}

func (ei EditInfoUD) IsDeleted() bool {
	return ei.DeletedAt == nil
}

func (ei EditInfoD) IsDeleted() bool {
	return ei.DeletedAt == nil
}

func CreateEditInfoCUD(createdBy *string) EditInfoCUD {
	var by string
	if createdBy != nil {
		by = *createdBy
	} else {
		by = system_user
	}
	return EditInfoCUD{
		EditInfoC: EditInfoC{
			CreatedBy: by,
			CreatedAt: time.Now().UTC(),
		},
	}
}

func CreateEditInfoCU(createdBy *string) EditInfoCU {
	var by string
	if createdBy != nil {
		by = *createdBy
	} else {
		by = system_user
	}
	return EditInfoCU{
		EditInfoC: EditInfoC{
			CreatedBy: by,
			CreatedAt: time.Now().UTC(),
		},
	}
}

func CreateEditInfoCD(createdBy *string) EditInfoCD {
	var by string
	if createdBy != nil {
		by = *createdBy
	} else {
		by = system_user
	}
	return EditInfoCD{
		EditInfoC: EditInfoC{
			CreatedBy: by,
			CreatedAt: time.Now().UTC(),
		},
	}
}

func CreateEditInfoC(createdBy *string) EditInfoC {
	var by string
	if createdBy != nil {
		by = *createdBy
	} else {
		by = system_user
	}
	return EditInfoC{
		CreatedBy: by,
		CreatedAt: time.Now().UTC(),
	}
}

type EditInfoCUD struct {
	EditInfoC
	EditInfoU
	EditInfoD
}

type EditInfoCU struct {
	EditInfoC
	EditInfoU
}

type EditInfoCD struct {
	EditInfoC
	EditInfoD
}

type EditInfoUD struct {
	EditInfoU
	EditInfoD
}

type EditInfoC struct {
	CreatedAt time.Time `column:"created_at"`
	CreatedBy string    `column:"created_by"`
}

type EditInfoU struct {
	UpdatedAt *time.Time `column:"updated_at"`
	UpdatedBy *string    `column:"updated_by"`
}

type EditInfoD struct {
	DeletedAt *time.Time `column:"deleted_at"`
	DeletedBy *string    `column:"deleted_by"`
}

func (ei *EditInfoU) PopulateUpdateFields(by *string) {
	updatedAt := time.Now().UTC()
	ei.UpdatedAt = &updatedAt
	if by != nil {
		ei.UpdatedBy = by
	} else {
		ei.UpdatedBy = &system_user
	}
}

func (ei *EditInfoD) PopulateDeleteFields(by *string) {
	deletedAt := time.Now().UTC()
	ei.DeletedAt = &deletedAt
	if by != nil {
		ei.DeletedBy = by
	} else {
		ei.DeletedBy = &system_user
	}
}

type EditInfoCUDSearchParams struct {
	EditInfoCSearchParams
	EditInfoUSearchParams
	EditInfoDSearchParams
}

func (sp *EditInfoCUDSearchParams) IsEmpty() bool {
	return sp.EditInfoCSearchParams.IsEmpty() && sp.EditInfoUSearchParams.IsEmpty() && sp.EditInfoDSearchParams.IsEmpty()
}

func (sp EditInfoCUDSearchParams) validate() error {
	err := sp.EditInfoCSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.EditInfoUSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.EditInfoDSearchParams.validate()
	if err != nil {
		return err
	}
	return nil
}

func (sp EditInfoCUDSearchParams) appendSearchQuery(tablePrefix string, query *string, params *[]interface{}) {
	sp.EditInfoCSearchParams.appendSearchQuery(tablePrefix, query, params)
	sp.EditInfoUSearchParams.appendSearchQuery(tablePrefix, query, params)
	sp.EditInfoDSearchParams.appendSearchQuery(tablePrefix, query, params)
}

type EditInfoCUSearchParams struct {
	EditInfoCSearchParams
	EditInfoUSearchParams
}

func (sp *EditInfoCUSearchParams) IsEmpty() bool {
	return sp.EditInfoCSearchParams.IsEmpty() && sp.EditInfoUSearchParams.IsEmpty()
}

func (sp EditInfoCUSearchParams) validate() error {
	err := sp.EditInfoCSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.EditInfoUSearchParams.validate()
	if err != nil {
		return err
	}
	return nil
}

func (sp EditInfoCUSearchParams) appendSearchQuery(tablePrefix string, query *string, params *[]interface{}) {
	sp.EditInfoCSearchParams.appendSearchQuery(tablePrefix, query, params)
	sp.EditInfoUSearchParams.appendSearchQuery(tablePrefix, query, params)
}

/*
type EditInfoCDSearchParams struct {
	EditInfoCSearchParams
	EditInfoDSearchParams
}

func (sp *EditInfoCDSearchParams) IsEmpty() bool {
	return sp.EditInfoCSearchParams.IsEmpty() && sp.EditInfoDSearchParams.IsEmpty()
}

func (sp EditInfoCDSearchParams) validate() error {
	err := sp.EditInfoCSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.EditInfoDSearchParams.validate()
	if err != nil {
		return err
	}
	return nil
}

func (sp EditInfoCDSearchParams) appendSearchQuery(tablePrefix string, query *string, params *[]interface{}) {
	sp.EditInfoCSearchParams.appendSearchQuery(tablePrefix, query, params)
	sp.EditInfoDSearchParams.appendSearchQuery(tablePrefix, query, params)
}
*/

type EditInfoCSearchParams struct {
	CreatedAtFrom   *time.Time `json:"createdAtFrom"`
	CreatedAtBefore *time.Time `json:"createdAtBefore"`
	CreatedBy       *string    `json:"createdBy"`
}

func (sp *EditInfoCSearchParams) IsEmpty() bool {
	return sp.CreatedAtFrom == nil && sp.CreatedAtBefore == nil && sp.CreatedBy == nil
}

func (sp EditInfoCSearchParams) validate() error {
	return nil
}

func (sp EditInfoCSearchParams) appendSearchQuery(tablePrefix string, query *string, params *[]interface{}) {
	if sp.CreatedAtFrom != nil && !sp.CreatedAtFrom.IsZero() {
		*params = append(*params, *sp.CreatedAtFrom)
		*query += fmt.Sprintf(" and %v.created_at>=$%d", tablePrefix, len(*params))
	}
	if sp.CreatedAtBefore != nil && !sp.CreatedAtBefore.IsZero() {
		*params = append(*params, *sp.CreatedAtBefore)
		*query += fmt.Sprintf(" and %v.created_at<$%d", tablePrefix, len(*params))
	}
	if sp.CreatedBy != nil && len(*sp.CreatedBy) != 0 {
		*params = append(*params, *sp.CreatedBy)
		*query += fmt.Sprintf(" and %v.created_by=$%d", tablePrefix, len(*params))
	}
}

type EditInfoUSearchParams struct {
	UpdatedAtFrom   *time.Time `json:"updatedAtFrom"`
	UpdatedAtBefore *time.Time `json:"updatedAtBefore"`
	UpdatedBy       *string    `json:"updatedBy"`
}

func (sp *EditInfoUSearchParams) IsEmpty() bool {
	return sp.UpdatedAtFrom == nil && sp.UpdatedAtBefore == nil && sp.UpdatedBy == nil
}

func (sp EditInfoUSearchParams) validate() error {
	return nil
}

func (sp EditInfoUSearchParams) appendSearchQuery(tablePrefix string, query *string, params *[]interface{}) {
	if sp.UpdatedAtFrom != nil && !sp.UpdatedAtFrom.IsZero() {
		*params = append(*params, *sp.UpdatedAtFrom)
		*query += fmt.Sprintf(" and %v.updated_at>=$%d", tablePrefix, len(*params))
	}
	if sp.UpdatedAtBefore != nil && !sp.UpdatedAtBefore.IsZero() {
		*params = append(*params, *sp.UpdatedAtBefore)
		*query += fmt.Sprintf(" and %v.updated_at<$%d", tablePrefix, len(*params))
	}
	if sp.UpdatedBy != nil && len(*sp.UpdatedBy) != 0 {
		*params = append(*params, *sp.UpdatedBy)
		*query += fmt.Sprintf(" and %v.updated_by=$%d", tablePrefix, len(*params))
	}
}

type EditInfoDSearchParams struct {
	DeletedAtFrom   *time.Time `json:"deletedAtFrom"`
	DeletedAtBefore *time.Time `json:"deletedAtToBefore"`
	DeletedBy       *string    `json:"deletedBy"`
}

func (sp *EditInfoDSearchParams) IsEmpty() bool {
	return sp.DeletedAtFrom == nil && sp.DeletedAtBefore == nil && sp.DeletedBy == nil
}

func (sp EditInfoDSearchParams) validate() error {
	return nil
}

func (sp EditInfoDSearchParams) appendSearchQuery(tablePrefix string, query *string, params *[]interface{}) {
	if sp.DeletedAtFrom != nil && !sp.DeletedAtFrom.IsZero() {
		*params = append(*params, *sp.DeletedAtFrom)
		*query += fmt.Sprintf(" and %v.deleted_at>=$%d", tablePrefix, len(*params))
	}
	if sp.DeletedAtBefore != nil && !sp.DeletedAtBefore.IsZero() {
		*params = append(*params, *sp.DeletedAtBefore)
		*query += fmt.Sprintf(" and %v.deleted_at<$%d", tablePrefix, len(*params))
	}
	if sp.DeletedBy != nil && len(*sp.DeletedBy) != 0 {
		*params = append(*params, *sp.DeletedBy)
		*query += fmt.Sprintf(" and %v.deleted_by=$%d", tablePrefix, len(*params))
	}
}

type PagingSearchParams struct {
	Limit  *int64 `json:"limit"`
	Offset *int64 `json:"offset"`
}

func (sp *PagingSearchParams) IsEmpty() bool {
	return sp.Limit == nil && sp.Offset == nil
}

func (sp PagingSearchParams) validate() error {
	return nil
}

func (sp *PagingSearchParams) SetLimit(limit *int64) {
	sp.Limit = limit
}

func (sp *PagingSearchParams) SetOffset(offset *int64) {
	sp.Offset = offset
}

func (sp PagingSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	if sp.Limit != nil && *sp.Limit >= 0 {
		*params = append(*params, *sp.Limit)
		*query += fmt.Sprintf(" limit $%d ", len(*params))
	}
	if sp.Offset != nil && *sp.Offset >= 0 {
		*params = append(*params, *sp.Offset)
		*query += fmt.Sprintf(" offset $%d ", len(*params))
	}
}

type SortColumn struct {
	Prefix    string
	Column    string
	Order     int
	Direction int
}

func (sc SortColumn) Clone() SortColumn {
	return SortColumn{
		Prefix:    sc.Prefix,
		Column:    sc.Column,
		Order:     sc.Order,
		Direction: sc.Direction,
	}
}

func (sc SortColumn) orderBy() string {
	dir := ""
	if sc.Direction < 0 {
		dir = " desc"
	}
	return sc.Prefix + "." + sc.Column + dir
}

type SortColumns []SortColumn

func (sc SortColumns) Len() int           { return len(sc) }
func (sc SortColumns) Less(i, j int) bool { return sc[i].Order < sc[j].Order }
func (sc SortColumns) Swap(i, j int)      { sc[i], sc[j] = sc[j], sc[i] }

func (sp SortColumns) OrderBy() string {
	sort.Sort(sp)
	result := ""
	for _, v := range sp {
		result += v.orderBy()
		result += ","
	}
	return result
}

type EditInfoCSortParams struct {
	Prefix    string
	CreatedAt *SortColumn `column:"created_at"`
	CreatedBy *SortColumn `column:"created_by"`
}

func (sp EditInfoCSortParams) IsEmpty() bool {
	return sp.CreatedAt == nil && sp.CreatedBy == nil
}

func (sp EditInfoCSortParams) SortColumns() SortColumns {
	var scs SortColumns
	if sp.CreatedAt != nil {
		sp.CreatedAt.Prefix = sp.Prefix
		sp.CreatedAt.Column = util.GetTag(sp, "CreatedAt", column_tag)
		scs = append(scs, *sp.CreatedAt)
	}
	if sp.CreatedBy != nil {
		sp.CreatedBy.Prefix = sp.Prefix
		sp.CreatedBy.Column = util.GetTag(sp, "CreatedBy", column_tag)
		scs = append(scs, *sp.CreatedBy)
	}
	return scs
}

type EditInfoUSortParams struct {
	Prefix    string
	UpdatedAt *SortColumn `column:"updated_at"`
	UpdatedBy *SortColumn `column:"updated_by"`
}

func (sp EditInfoUSortParams) IsEmpty() bool {
	return sp.UpdatedAt == nil && sp.UpdatedBy == nil
}

func (sp EditInfoUSortParams) SortColumns() SortColumns {
	var scs SortColumns
	if sp.UpdatedAt != nil {
		sp.UpdatedAt.Prefix = sp.Prefix
		sp.UpdatedAt.Column = util.GetTag(sp, "UpdatedAt", column_tag)
		scs = append(scs, *sp.UpdatedAt)
	}
	if sp.UpdatedBy != nil {
		sp.UpdatedBy.Prefix = sp.Prefix
		sp.UpdatedBy.Column = util.GetTag(sp, "UpdatedBy", column_tag)
		scs = append(scs, *sp.UpdatedBy)
	}
	return scs
}

type EditInfoDSortParams struct {
	Prefix    string
	DeletedAt *SortColumn `column:"deleted_at"`
	DeletedBy *SortColumn `column:"deleted_by"`
}

func (sp EditInfoDSortParams) IsEmpty() bool {
	return sp.DeletedAt == nil && sp.DeletedBy == nil
}

func (sp EditInfoDSortParams) SortColumns() SortColumns {
	var scs SortColumns
	if sp.DeletedAt != nil {
		sp.DeletedAt.Prefix = sp.Prefix
		sp.DeletedAt.Column = util.GetTag(sp, "DeletedAt", column_tag)
		scs = append(scs, *sp.DeletedAt)
	}
	if sp.DeletedBy != nil {
		sp.DeletedBy.Prefix = sp.Prefix
		sp.DeletedBy.Column = util.GetTag(sp, "DeletedBy", column_tag)
		scs = append(scs, *sp.DeletedBy)
	}
	return scs
}

type EditInfoCUSortParams struct {
	Prefix string
	EditInfoCSortParams
	EditInfoUSortParams
}

func (sp EditInfoCUSortParams) IsEmpty() bool {
	return sp.EditInfoCSortParams.IsEmpty() && sp.EditInfoUSortParams.IsEmpty()
}

func (sp EditInfoCUSortParams) SortColumns() SortColumns {
	var scs SortColumns
	if !sp.EditInfoCSortParams.IsEmpty() {
		sp.EditInfoCSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoCSortParams.SortColumns()...)
	}
	if !sp.EditInfoUSortParams.IsEmpty() {
		sp.EditInfoUSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoUSortParams.SortColumns()...)
	}
	return scs
}

type EditInfoCDSortParams struct {
	Prefix string
	EditInfoCSortParams
	EditInfoDSortParams
}

func (sp EditInfoCDSortParams) IsEmpty() bool {
	return sp.EditInfoCSortParams.IsEmpty() && sp.EditInfoDSortParams.IsEmpty()
}

func (sp EditInfoCDSortParams) SortColumns() SortColumns {
	var scs SortColumns
	if !sp.EditInfoCSortParams.IsEmpty() {
		sp.EditInfoCSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoCSortParams.SortColumns()...)
	}
	if !sp.EditInfoDSortParams.IsEmpty() {
		sp.EditInfoDSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoDSortParams.SortColumns()...)
	}
	return scs
}

type EditInfoCUDSortParams struct {
	Prefix string
	EditInfoCSortParams
	EditInfoUSortParams
	EditInfoDSortParams
}

func (sp EditInfoCUDSortParams) IsEmpty() bool {
	return sp.EditInfoCSortParams.IsEmpty() && sp.EditInfoUSortParams.IsEmpty() && sp.EditInfoDSortParams.IsEmpty()
}

func (sp EditInfoCUDSortParams) SortColumns() SortColumns {
	var scs SortColumns
	if !sp.EditInfoCSortParams.IsEmpty() {
		sp.EditInfoCSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoCSortParams.SortColumns()...)
	}
	if !sp.EditInfoUSortParams.IsEmpty() {
		sp.EditInfoUSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoUSortParams.SortColumns()...)
	}
	if !sp.EditInfoDSortParams.IsEmpty() {
		sp.EditInfoDSortParams.Prefix = sp.Prefix
		scs = append(scs, sp.EditInfoDSortParams.SortColumns()...)
	}
	return scs
}

func ParseSortParams(sort *string) SortColumns {
	res := SortColumns{}
	if sort != nil {
		split := strings.Split(*sort, ",")
		for i, s := range split {
			if len(s) > 0 {
				first := s[0:1]
				var direction int
				var rest string
				switch first {
				case "+":
					direction = 1
					rest = s[1:]
				case "-":
					direction = -1
					rest = s[1:]
				default:
					direction = 1
					rest = s
				}
				cs := SortColumn{
					Column:    rest,
					Order:     i,
					Direction: direction,
				}
				res = append(res, cs)
			}
		}
	}
	return res
}

type GroupColumn struct {
	Prefix string
	Column string
	Order  int
}

func (gc GroupColumn) Clone() SortColumn {
	return SortColumn{
		Prefix: gc.Prefix,
		Column: gc.Column,
		Order:  gc.Order,
	}
}

func (gc GroupColumn) groupBy() string {
	return gc.Prefix + "." + gc.Column
}

type GroupColumns []GroupColumn

func (gc GroupColumns) Len() int           { return len(gc) }
func (gc GroupColumns) Less(i, j int) bool { return gc[i].Order < gc[j].Order }
func (gc GroupColumns) Swap(i, j int)      { gc[i], gc[j] = gc[j], gc[i] }

func (gc GroupColumns) GroupBy() string {
	sort.Sort(gc)
	result := ""
	for _, v := range gc {
		result += v.groupBy()
		result += ","
	}
	return result
}

func AppendQuery(sp SearchParams, query *string, params *[]interface{}) error {
	err := sp.validate()
	if err != nil {
		return err
	}
	sp.joinTables(query)
	*query += "\n"
	sp.appendSearchQuery(query, params)
	*query += "\n"
	sp.appendSortQuery(query)
	if (*query)[len(*query)-1] == ',' {
		*query = (*query)[0 : len(*query)-1]
	}
	*query += "\n"
	sp.appendGroupByQuery(query)
	if (*query)[len(*query)-1] == ',' {
		*query = (*query)[0 : len(*query)-1]
	}
	*query += "\n"
	sp.appendPagingQuery(query, params)
	return nil
}

func AppendCountQuery(sp SearchParams, query *string, params *[]interface{}) error {
	err := sp.validate()
	if err != nil {
		return err
	}
	sp.joinTables(query)
	*query += "\n"
	sp.appendSearchQuery(query, params)
	return nil
}

func AppendUpdateQuery(up UpdateParams, query *string, params *[]interface{}) {
	up.appendUpdateQuery(query, params)
}

type EditInfoUDUpdateParams struct {
	EditInfoUD
}

func (up EditInfoUDUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*params = append(*params, up.UpdatedAt)
	*query += fmt.Sprintf("updated_at = $%d, ", len(*params))
	*params = append(*params, up.UpdatedBy)
	*query += fmt.Sprintf("updated_by = $%d ", len(*params))
	if up.DeletedAt != nil {
		*params = append(*params, up.DeletedAt)
		*query += fmt.Sprintf(", deleted_at = $%d, ", len(*params))
		*params = append(*params, up.DeletedBy)
		*query += fmt.Sprintf("deleted_by = $%d ", len(*params))
	}
}

type EditInfoUUpdateParams struct {
	EditInfoU
}

func (up EditInfoUUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*params = append(*params, up.UpdatedAt)
	*query += fmt.Sprintf("updated_at = $%d, ", len(*params))
	*params = append(*params, up.UpdatedBy)
	*query += fmt.Sprintf("updated_by = $%d ", len(*params))
}

type Review struct {
	Grade   float64 `json:"grade,omitempty"`
	Comment string  `json:"comment,omitempty"`
	UserId  string  `json:"userId,omitempty"`
}

type Reviews struct {
	Reviews []Review `json:"reviews,omitempty"`
	Average float64  `json:"average"`
}

// Value is implementation of data Valuer interface.
func (r Reviews) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan is implementation of database/sql scanner interface.
func (r *Reviews) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}

type Apointment struct {
	StartTime  time.Time `json:"startTime,omitempty"`
	EndTime    time.Time `json:"endTime,omitempty"`
	Accepted   bool      `json:"accepted,omitempty"`
	PracticeId string    `json:"id,omitempty"`
}

type Booking []Apointment

// Value is implementation of data Valuer interface.
func (bo Booking) Value() (driver.Value, error) {
	return json.Marshal(bo)
}

// Scan is implementation of database/sql scanner interface.
func (bo *Booking) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &bo)
}
