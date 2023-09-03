package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type ApiJournalCrud struct {
	Crud
}

func InitApiJournalCrud(db *sql.DB) *ApiJournalCrud {
	return &ApiJournalCrud{
		Crud{
			db: db,
		},
	}
}

const (
	api_journal_select = `
		select aj.api_journal_id, aj.user_id, aj.request, aj.response, aj.request_json, aj.response_json, aj.created_at, aj.created_by, aj.updated_at, aj.updated_by, aj.source_ip
		from api_journal aj
	`
	api_journal_count = `select count(*) from api_journal aj `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *ApiJournalCrud) checkConstraints(ctx context.Context, en DR.ApiJournal, qa QueryAble) error {
	return nil
}

func (r *ApiJournalCrud) populateFKFields(ctx context.Context, en *DR.ApiJournal, qa QueryAble) {
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a api_journal_journal
func (r *ApiJournalCrud) Create(ctx context.Context, en DR.ApiJournal, qa QueryAble, by *string) (DR.ApiJournal, error) {
	L.L.WithRequestID(ctx).Info("ApiJournalCrud.Create", L.Any("apiJournal", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	// checking constraints
	err := r.checkConstraints(ctx, en, qa)
	if err != nil {
		return en, err
	}

	var reqJson *string
	if en.RequestBodyString != nil && util.IsJSON(*en.RequestBodyString) {
		reqJson = en.RequestBodyString
	}

	query := `insert into api_journal (request, request_json, created_at, created_by,source_ip)
	values ($1, $2, $3, $4, $5) RETURNING api_journal_id;`
	params := []interface{}{en.Request, reqJson, en.CreatedAt, en.CreatedBy, en.SourceIP}

	L.L.Debug("ApiJournalJournalCrud.Create insert", L.String("query", query), L.Any("params", params))

	err = db.QueryRowContext(ctx, query, params...).Scan(&en.ApiJournalId)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.ApiJournalId, qa)
	if err != nil {
		util.LogPqError(ctx, err)
		return pen, err
	}

	return pen, nil
}

////////////////////////////////////////////////READ/////////////////////////////////////////////////////////////////////////////////////

// GetById returns api_journal_journal by id
func (r *ApiJournalCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.ApiJournal, error) {
	L.L.WithRequestID(ctx).Info("ApiJournalCrud.GetById", L.String("apiJournalId", id))

	db := r.GetTx(qa)

	aj := DR.ApiJournal{}
	query := ""
	if qa != nil {
		query = api_journal_select +
			`where aj.api_journal_id=$1 for update`
	} else {
		query = api_journal_select +
			`where aj.api_journal_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&aj.ApiJournalId, &aj.UserId, &aj.Request, &aj.Response, &aj.RequestJson, &aj.ResponseJson, &aj.CreatedAt, &aj.CreatedBy, &aj.UpdatedAt, &aj.UpdatedBy, &aj.SourceIP)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("api journal does not exist for id: %v", id)
		}
	} else {
		r.populateFKFields(ctx, &aj, qa)
	}
	return aj, err
}

func (r *ApiJournalCrud) GetCount(ctx context.Context, sp DR.ApiJournalSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("ApiJournalCrud.GetCount", L.Any("apiJournal", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := api_journal_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("ApiJournalCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *ApiJournalCrud) Search(ctx context.Context, sp DR.ApiJournalSearchParams, qa QueryAble) ([]DR.ApiJournal, error) {
	L.L.WithRequestID(ctx).Info("ApiJournalCrud.Search", L.Any("apiJournal", sp))

	db := r.GetTx(qa)

	results := []DR.ApiJournal{}
	var params []interface{}

	query := api_journal_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("ApiJournalCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		aj := DR.ApiJournal{}
		err := rows.Scan(&aj.ApiJournalId, &aj.UserId, &aj.Request, &aj.Response, &aj.RequestJson, &aj.ResponseJson, &aj.CreatedAt, &aj.CreatedBy, &aj.UpdatedAt, &aj.UpdatedBy, &aj.SourceIP)
		if err != nil {
			return nil, err
		}
		r.populateFKFields(ctx, &aj, qa)
		results = append(results, aj)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("ApiJournalCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a api_journal
func (r *ApiJournalCrud) Update(ctx context.Context, up DR.ApiJournalUpdateParams, qa QueryAble, by *string) (DR.ApiJournal, error) {
	L.L.WithRequestID(ctx).Info("ApiJournalCrud.Update", L.Any("ApiJournal", up))

	up.PopulateUpdateFields(by)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("ApiJournalCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.ApiJournal{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.ApiJournal{}, fmt.Errorf("no rows affected")
	}
	pen, err := r.GetById(ctx, up.Id, qa)
	if err != nil {
		util.LogPqError(ctx, err)
		return pen, err
	}

	return pen, nil
}
