package crud

import (
	L "backend/internal/logging"
	"backend/sportos"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type AuditCrud struct {
	Crud
	Enabled bool
}

func InitAuditCrud(db *sql.DB) *AuditCrud {
	return &AuditCrud{
		Crud{
			db: db,
		},
		false,
	}
}

func (r *AuditCrud) Start() {
	r.Enabled = true
}

func (r *AuditCrud) Stop() {
	r.Enabled = false
}

const (
	audit_select = `
		select aud.audit_id, aud.entity, aud.entity_id, aud.crud_action, aud.old, aud.new, aud.api_journal_id, aud.created_at, aud.created_by
		from audit aud
	`
	audit_count = `select count(*) from audit aud `
)

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates an audit from old and new value of entity that was affected by CRUD operation
func (r *AuditCrud) CreateSnapshot(ctx context.Context, old, new DR.CommonEntity, qa QueryAble, by *string) (DR.Audit, error) {
	if !r.Enabled {
		return DR.Audit{}, nil
	}
	var auditNew = make(map[string]interface{})
	var auditOld = make(map[string]interface{})
	var id string
	var name DR.SportosEntity
	var op DR.CrudAction
	var convertErr error
	if new != nil {
		op = DR.AUDIT_CREATE
		name = new.GetTableName()
		id = new.GetId()
	}
	if old != nil {
		if op != "" {
			op = DR.AUDIT_UPDATE
		} else {
			op = DR.AUDIT_DELETE
		}
		name = old.GetTableName()
		id = old.GetId()
	}
	switch op {
	case DR.AUDIT_CREATE:
		auditOld = nil
		convertErr = util.ConvertStructToJSONHash(auditNew, new, "column")
		if convertErr != nil {
			return DR.Audit{}, convertErr
		}
	case DR.AUDIT_DELETE:
		auditNew = nil
		convertErr = util.ConvertStructToJSONHash(auditOld, old, "column")
		if convertErr != nil {
			return DR.Audit{}, convertErr
		}
	case DR.AUDIT_UPDATE:
		oldDiff, newDiff, err := util.GenerateDiferences(old, new)
		if err != nil {
			return DR.Audit{}, err
		}
		convertErr = util.ConvertStructToJSONHash(auditOld, oldDiff, "column")
		if convertErr != nil {
			return DR.Audit{}, convertErr
		}
		convertErr = util.ConvertStructToJSONHash(auditNew, newDiff, "column")
		if convertErr != nil {
			return DR.Audit{}, convertErr
		}

	}
	var apiJournalId *string = nil
	if val, ok := ctx.Value(sportos.CONTEXT_API_JOURNAL_ID_KEY).(string); ok {
		apiJournalId = &val
	}
	auditVal := DR.Audit{
		Entity:       name,
		EntityId:     id,
		CrudAction:   &op,
		ApiJournalId: apiJournalId,
		Old:          auditOld,
		New:          auditNew,
	}
	ret, err := DR.Audit{}, error(nil)
	if len(auditVal.Old) > 0 || len(auditVal.New) > 0 {
		ret, err = r.Create(ctx, auditVal, qa, by)
	}
	return ret, err
}

// Creates an audit
func (r *AuditCrud) Create(ctx context.Context, en DR.Audit, qa QueryAble, by *string) (DR.Audit, error) {
	L.L.WithRequestID(ctx).Info("AuditCrud.Create", L.Any("audit", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	query := `insert into audit (entity, entity_id, crud_action, old, new, api_journal_id, created_by, created_at) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING audit_id;`
	params := []interface{}{en.Entity, en.EntityId, en.CrudAction, en.Old, en.New, en.ApiJournalId, en.CreatedBy, en.CreatedAt}

	L.L.Debug("AuditCrud.Create insert", L.String("query", query), L.Any("params", params))

	err := db.QueryRowContext(ctx, query, params...).Scan(&en.AuditId)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.AuditId, qa)
	if err != nil {
		util.LogPqError(ctx, err)
		return pen, err
	}

	return pen, nil
}

////////////////////////////////////////////////READ///////////////////////////////////////////////////////////////////////////////////

// Returns audit by id
func (r *AuditCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Audit, error) {
	L.L.WithRequestID(ctx).Info("AuditCrud.getById", L.String("auditId", id))

	db := r.GetTx(qa)

	c := DR.Audit{}
	query := ""
	if qa != nil {
		query = audit_select +
			`where aud.audit_id=$1 for update`
	} else {
		query = audit_select +
			`where aud.audit_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&c.AuditId, &c.Entity, &c.EntityId, &c.CrudAction, &c.Old, &c.New, &c.ApiJournalId, &c.CreatedAt, &c.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("audit does not exist for id: %v", id)
		}
	}

	return c, err
}

func (r *AuditCrud) GetCount(ctx context.Context, sp DR.AuditSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("AuditCrud.GetCount", L.Any("audit", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := audit_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("AuditCrud.GetCount query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return 0, err
	}
	defer rows.Close()

	cnt := 0
	for rows.Next() {
		err := rows.Scan(&cnt)
		if err != nil {
			return 0, err
		}
	}
	return cnt, nil
}

func (r *AuditCrud) Search(ctx context.Context, sp DR.AuditSearchParams, qa QueryAble) ([]DR.Audit, error) {

	L.L.WithRequestID(ctx).Info("AuditCrud.Search", L.Any("auditSearchParams", sp))

	db := r.GetTx(qa)

	results := []DR.Audit{}
	var params []interface{}

	query := audit_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("AuditCrud.Search query", L.String("query", query), L.Any("params", params))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		audit := DR.Audit{}
		err := rows.Scan(&audit.AuditId, &audit.Entity, &audit.EntityId, &audit.CrudAction, &audit.Old, &audit.New, &audit.ApiJournalId, &audit.CreatedAt, &audit.CreatedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, audit)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("AuditCrud.Search No rows returned ")
	}
	return results, nil
}
