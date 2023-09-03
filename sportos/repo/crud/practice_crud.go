package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type PracticeCrud struct {
	Crud
}

func InitPracticeCrud(db *sql.DB) *PracticeCrud {
	return &PracticeCrud{
		Crud{
			db: db,
		},
	}
}

const (
	practice_select = `
		select pr.practice_id, pr.player_id, pr.coach_id, pr.status, pr.start_time, pr.sport, pr.created_at, pr.created_by, pr.updated_at, pr.updated_by, pr.deleted_at, pr.deleted_by
		from practice pr
	`
	practice_count = `select count(*) from practice pr `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a Practice
func (r *PracticeCrud) Create(ctx context.Context, en DR.Practice, qa QueryAble, by *string) (DR.Practice, error) {
	L.L.WithRequestID(ctx).Info("PracticeCrud.Create", L.Any("practice", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	query := `insert into practice (player_id, coach_id, start_time, status, sport, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6, $7) RETURNING practice_id;`
	params := []interface{}{en.PlayerId, en.CoachId, en.StartTime, en.Status, en.Sport, en.CreatedAt, en.CreatedBy}

	L.L.Debug("PracticeCrud.Create insert", L.String("query", query), L.Any("params", params))

	err := db.QueryRowContext(ctx, query, params...).Scan(&en.PracticeId)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.PracticeId, qa)
	if err != nil {
		util.LogPqError(ctx, err)
		return pen, err
	}

	_, err = r.crudRepo.AuditCrud.CreateSnapshot(ctx, nil, &pen, qa, by)
	if err != nil {
		return pen, err
	}

	return pen, nil
}

////////////////////////////////////////////////READ/////////////////////////////////////////////////////////////////////////////////////

// GetById returns practice by id
func (r *PracticeCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Practice, error) {
	L.L.WithRequestID(ctx).Info("PracticeCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	pr := DR.Practice{}
	query := ""
	if qa != nil {
		query = practice_select +
			`where pr.practice_id=$1 for update`
	} else {
		query = practice_select +
			`where pr.practice_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&pr.PracticeId, &pr.PlayerId, &pr.CoachId, &pr.Status, &pr.StartTime, &pr.Sport, &pr.CreatedAt, &pr.CreatedBy, &pr.UpdatedAt, &pr.UpdatedBy, &pr.DeletedAt, &pr.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("practice does not exist for username: %v", id)
		}
	}

	return pr, err
}

func (r *PracticeCrud) GetCount(ctx context.Context, sp DR.PracticeSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("PracticeCrud.GetCount", L.Any("practice", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := practice_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("PracticeCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *PracticeCrud) Search(ctx context.Context, sp DR.PracticeSearchParams, qa QueryAble) ([]DR.Practice, error) {
	L.L.WithRequestID(ctx).Info("PracticeCrud.Search", L.Any("practice", sp))

	db := r.GetTx(qa)

	results := []DR.Practice{}
	var params []interface{}

	query := practice_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("PracticeCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		pr := DR.Practice{}
		err := rows.Scan(&pr.PracticeId, &pr.PlayerId, &pr.CoachId, &pr.Status, &pr.StartTime, &pr.Sport, &pr.CreatedAt, &pr.CreatedBy, &pr.UpdatedAt, &pr.UpdatedBy, &pr.DeletedAt, &pr.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, pr)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("PracticeCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a practice
func (r *PracticeCrud) Update(ctx context.Context, up DR.PracticeUpdateParams, qa QueryAble, by *string) (DR.Practice, error) {
	L.L.WithRequestID(ctx).Info("PracticeCrud.Update", L.Any("practice", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("PracticeCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Practice{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Practice{}, fmt.Errorf("no rows affected")
	}
	pen, err := r.GetById(ctx, up.Id, qa)
	if err != nil {
		util.LogPqError(ctx, err)
		return pen, err
	}

	_, err = r.crudRepo.AuditCrud.CreateSnapshot(ctx, &old, &pen, qa, by)
	if err != nil {
		return pen, err
	}

	return pen, nil
}
