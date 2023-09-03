package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type CoachCrud struct {
	Crud
}

func InitCoachCrud(db *sql.DB) *CoachCrud {
	return &CoachCrud{
		Crud{
			db: db,
		},
	}
}

const (
	coach_select = `
		select co.user_id, co.name, co.city, co.sport, co.reviews, co.booking, co.created_at, co.created_by, co.updated_at, co.updated_by, co.deleted_at, co.deleted_by
		from coach co
	`
	coach_count = `select count(*) from coach co `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *CoachCrud) exists(ctx context.Context, id string, qa QueryAble) bool {
	L.L.WithRequestID(ctx).Info("CoachCrud.exists", L.Any("id", id))

	db := r.GetTx(qa)

	var count int
	row := db.QueryRowContext(ctx, `select count(*) from coach where user_id=$1`, id)

	err := row.Scan(&count)
	if err != nil {
		L.L.Error("CoachCrud.exists error", L.Any("err", err))
	}

	return count > 0
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a Coach
func (r *CoachCrud) Create(ctx context.Context, en DR.Coach, qa QueryAble, by *string) (DR.Coach, error) {
	L.L.WithRequestID(ctx).Info("CoachCrud.Create", L.Any("coach", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	if r.exists(ctx, en.Username, qa) {
		return en, fmt.Errorf("user with username %s already exists", en.Username)
	}

	query := `insert into coach (user_id, name, city, sport, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6) RETURNING user_id;`
	params := []interface{}{en.Username, en.Name, en.City, en.Sport, en.CreatedAt, en.CreatedBy}

	L.L.Debug("CoachCrud.Create insert", L.String("query", query), L.Any("params", params))

	err := db.QueryRowContext(ctx, query, params...).Scan(&en.Username)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.Username, qa)
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

// GetById returns coach by userrname
func (r *CoachCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Coach, error) {
	L.L.WithRequestID(ctx).Info("CoachCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	co := DR.Coach{}
	query := ""
	if qa != nil {
		query = coach_select +
			`where co.user_id=$1 for update`
	} else {
		query = coach_select +
			`where co.user_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&co.Username, &co.Name, &co.City, &co.Sport, &co.Reviews, &co.Booking, &co.CreatedAt, &co.CreatedBy, &co.UpdatedAt, &co.UpdatedBy, &co.DeletedAt, &co.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("coach does not exist for username: %v", id)
		}
	}

	return co, err
}

// GetById returns coach by email
func (r *CoachCrud) GetByEmail(ctx context.Context, email string, qa QueryAble) (DR.Coach, error) {
	L.L.WithRequestID(ctx).Info("CoachCrud.GetById", L.String("email", email))

	db := r.GetTx(qa)

	co := DR.Coach{}
	query := ""
	if qa != nil {
		query = coach_select +
			`where co.email=$1 for update`
	} else {
		query = coach_select +
			`where co.email=$1`
	}
	row := db.QueryRowContext(ctx, query,
		email)

	err := row.Scan(&co.Username, &co.Name, &co.City, &co.Sport, &co.Reviews, &co.Booking, &co.CreatedAt, &co.CreatedBy, &co.UpdatedAt, &co.UpdatedBy, &co.DeletedAt, &co.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("coach does not exist for email: %v", email)
		}
	}

	return co, err
}

func (r *CoachCrud) GetCount(ctx context.Context, sp DR.CoachSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("CoachCrud.GetCount", L.Any("coach", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := coach_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("CoachCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *CoachCrud) Search(ctx context.Context, sp DR.CoachSearchParams, qa QueryAble) ([]DR.Coach, error) {
	L.L.WithRequestID(ctx).Info("CoachCrud.Search", L.Any("coach", sp))

	db := r.GetTx(qa)

	results := []DR.Coach{}
	var params []interface{}

	query := coach_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("CoachCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		co := DR.Coach{}
		err := rows.Scan(&co.Username, &co.Name, &co.City, &co.Sport, &co.Reviews, &co.Booking, &co.CreatedAt, &co.CreatedBy, &co.UpdatedAt, &co.UpdatedBy, &co.DeletedAt, &co.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, co)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("CoachCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a coach
func (r *CoachCrud) Update(ctx context.Context, up DR.CoachUpdateParams, qa QueryAble, by *string) (DR.Coach, error) {
	L.L.WithRequestID(ctx).Info("CoachCrud.Update", L.Any("coach", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("CoachCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Coach{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Coach{}, fmt.Errorf("no rows affected")
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
