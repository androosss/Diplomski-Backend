package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type PlaceCrud struct {
	Crud
}

func InitPlaceCrud(db *sql.DB) *PlaceCrud {
	return &PlaceCrud{
		Crud{
			db: db,
		},
	}
}

const (
	place_select = `
		select pla.user_id, pla.name, pla.city, pla.sport, pla.reviews, pla.booking, pla.created_at, pla.created_by, pla.updated_at, pla.updated_by, pla.deleted_at, pla.deleted_by
		from place pla
	`
	place_count = `select count(*) from place pla `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *PlaceCrud) exists(ctx context.Context, id string, qa QueryAble) bool {
	L.L.WithRequestID(ctx).Info("PlaceCrud.exists", L.Any("id", id))

	db := r.GetTx(qa)

	var count int
	row := db.QueryRowContext(ctx, `select count(*) from place where user_id=$1`, id)

	err := row.Scan(&count)
	if err != nil {
		L.L.Error("PlaceCrud.exists error", L.Any("err", err))
	}

	return count > 0
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a Place
func (r *PlaceCrud) Create(ctx context.Context, en DR.Place, qa QueryAble, by *string) (DR.Place, error) {
	L.L.WithRequestID(ctx).Info("PlaceCrud.Create", L.Any("place", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	if r.exists(ctx, en.Username, qa) {
		return en, fmt.Errorf("user with username %s already exists", en.Username)
	}

	query := `insert into place (user_id, name, city, sport, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6) RETURNING user_id;`
	params := []interface{}{en.Username, en.Name, en.City, en.Sport, en.CreatedAt, en.CreatedBy}

	L.L.Debug("PlaceCrud.Create insert", L.String("query", query), L.Any("params", params))

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

// GetById returns place by userrname
func (r *PlaceCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Place, error) {
	L.L.WithRequestID(ctx).Info("PlaceCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	pla := DR.Place{}
	query := ""
	if qa != nil {
		query = place_select +
			`where pla.user_id=$1 for update`
	} else {
		query = place_select +
			`where pla.user_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&pla.Username, &pla.Name, &pla.City, &pla.Sport, &pla.Reviews, &pla.Booking, &pla.CreatedAt, &pla.CreatedBy, &pla.UpdatedAt, &pla.UpdatedBy, &pla.DeletedAt, &pla.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("place does not exist for username: %v", id)
		}
	}

	return pla, err
}

// GetById returns place by email
func (r *PlaceCrud) GetByEmail(ctx context.Context, email string, qa QueryAble) (DR.Place, error) {
	L.L.WithRequestID(ctx).Info("PlaceCrud.GetById", L.String("email", email))

	db := r.GetTx(qa)

	pla := DR.Place{}
	query := ""
	if qa != nil {
		query = place_select +
			`where pla.email=$1 for update`
	} else {
		query = place_select +
			`where pla.email=$1`
	}
	row := db.QueryRowContext(ctx, query,
		email)

	err := row.Scan(&pla.Username, &pla.Name, &pla.City, &pla.Sport, &pla.Reviews, &pla.Booking, &pla.CreatedAt, &pla.CreatedBy, &pla.UpdatedAt, &pla.UpdatedBy, &pla.DeletedAt, &pla.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("place does not exist for email: %v", email)
		}
	}

	return pla, err
}

func (r *PlaceCrud) GetCount(ctx context.Context, sp DR.PlaceSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("PlaceCrud.GetCount", L.Any("place", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := place_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("PlaceCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *PlaceCrud) Search(ctx context.Context, sp DR.PlaceSearchParams, qa QueryAble) ([]DR.Place, error) {
	L.L.WithRequestID(ctx).Info("PlaceCrud.Search", L.Any("place", sp))

	db := r.GetTx(qa)

	results := []DR.Place{}
	var params []interface{}

	query := place_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("PlaceCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		pla := DR.Place{}
		err := rows.Scan(&pla.Username, &pla.Name, &pla.City, &pla.Sport, &pla.Reviews, &pla.Booking, &pla.CreatedAt, &pla.CreatedBy, &pla.UpdatedAt, &pla.UpdatedBy, &pla.DeletedAt, &pla.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, pla)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("PlaceCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a place
func (r *PlaceCrud) Update(ctx context.Context, up DR.PlaceUpdateParams, qa QueryAble, by *string) (DR.Place, error) {
	L.L.WithRequestID(ctx).Info("PlaceCrud.Update", L.Any("place", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("PlaceCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Place{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Place{}, fmt.Errorf("no rows affected")
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
