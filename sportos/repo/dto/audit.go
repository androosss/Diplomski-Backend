package dto

import (
	"backend/sportos/repo/util"
	"fmt"
	"strings"
)

// [swagger]

// CrudAction
//
// Crud action. Possible values:
//   - `CREATE`
//   - `UPDATE`
//   - `DELETE`
//
// swagger:model CrudAction
type CrudAction string

var (
	AUDIT_CREATE CrudAction = "CREATE"
	AUDIT_UPDATE CrudAction = "UPDATE"
	AUDIT_DELETE CrudAction = "DELETE"
)

type Audit struct {
	AuditId      string        `json:"auditId"`
	Entity       SportosEntity `json:"entity"`
	EntityId     string        `json:"entityId"`
	CrudAction   *CrudAction   `json:"crudAction"`
	Old          UntypedConfig `json:"old"`
	New          UntypedConfig `json:"new"`
	ApiJournalId *string       `json:"apiJournalId"`
	EditInfoC
}

type AuditSearchParams struct {
	Entity     *SportosEntity
	EntityId   *string
	CrudAction *string
	SourceIp   *string
	EditInfoCSearchParams
	AuditSortParams
	PagingSearchParams
	prefix string
}

func (sp *AuditSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "aud"
}

func (sp *AuditSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.AuditSortParams.SetTablePrefix(prefix)
}

func (sp *AuditSearchParams) validate() error {
	err := sp.EditInfoCSearchParams.validate()
	if err != nil {
		return err
	}
	/*err=sp.PagingSearchParams.validate()
	if err!=nil {
		return err
	}*/
	return nil
}

func (sp *AuditSearchParams) joinTables(query *string) {
}

func (sp *AuditSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	//audit params
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	tablePrefix := sp.GetTablePrefix()
	if sp.Entity != nil && len(*sp.Entity) != 0 {
		*params = append(*params, *sp.Entity)
		*query += fmt.Sprintf(" and %v.entity=$%d", tablePrefix, len(*params))
	}
	if sp.EntityId != nil && len(*sp.EntityId) != 0 {
		*params = append(*params, *sp.EntityId)
		*query += fmt.Sprintf(" and %v.entity_id=$%d", tablePrefix, len(*params))
	}
	if sp.CrudAction != nil && len(*sp.CrudAction) != 0 {
		*params = append(*params, *sp.CrudAction)
		*query += fmt.Sprintf(" and %v.crud_action=$%d", tablePrefix, len(*params))
	}
	if sp.SourceIp != nil && len(*sp.SourceIp) != 0 {
		*params = append(*params, *sp.SourceIp)
		*query += fmt.Sprintf(" and %v.source_ip=$%d", tablePrefix, len(*params))
	}
	if !sp.EditInfoCSearchParams.IsEmpty() {
		sp.EditInfoCSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
}

func (sp *AuditSearchParams) appendSortQuery(query *string) {
	if !sp.AuditSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.AuditSortParams.OrderBy()
	}
}

func (sp *AuditSearchParams) appendGroupByQuery(query *string) {

}

func (sp *AuditSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type AuditSortParams struct {
	Prefix   string
	EntityId *SortColumn `column:"entity_id"`
	Entity   *SortColumn `column:"entity"`
	EditInfoCUDSortParams
}

func (sp AuditSortParams) IsEmpty() bool {
	return sp.EntityId == nil && sp.Entity == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp AuditSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "aud"
}

func (sp *AuditSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp AuditSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.EntityId != nil {
		sp.EntityId.Prefix = tablePrefix
		sp.EntityId.Column = util.GetTag(sp, "EntityId", column_tag)
		scs = append(scs, *sp.EntityId)
	}
	if sp.Entity != nil {
		sp.Entity.Prefix = tablePrefix
		sp.Entity.Column = util.GetTag(sp, "Entity", column_tag)
		scs = append(scs, *sp.Entity)
	}
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp AuditSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}
