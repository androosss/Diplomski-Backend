package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type MatchCrud struct {
	Crud
}

func InitMatchCrud(db *sql.DB) *MatchCrud {
	return &MatchCrud{
		Crud{
			db: db,
		},
	}
}

const (
	match_select = `
		select ma.match_id, ma.status, ma.start_time, ma.players, ma.result, ma.place_id, ma.sport, ma.teams, ma.created_at, ma.created_by, ma.updated_at, ma.updated_by, ma.deleted_at, ma.deleted_by
		from match ma
	`
	match_count = `select count(*) from match ma `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a Match
func (r *MatchCrud) Create(ctx context.Context, en DR.Match, qa QueryAble, by *string) (DR.Match, error) {
	L.L.WithRequestID(ctx).Info("MatchCrud.Create", L.Any("match", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	query := `insert into match (start_time, place_id, status, players, sport, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6, $7) RETURNING match_id;`
	params := []interface{}{en.StartTime, en.PlaceId, en.Status, en.Players, en.Sport, en.CreatedAt, en.CreatedBy}

	L.L.Debug("MatchCrud.Create insert", L.String("query", query), L.Any("params", params))

	err := db.QueryRowContext(ctx, query, params...).Scan(&en.MatchId)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.MatchId, qa)
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

// GetById returns match by id
func (r *MatchCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Match, error) {
	L.L.WithRequestID(ctx).Info("MatchCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	ma := DR.Match{}
	query := ""
	if qa != nil {
		query = match_select +
			`where ma.match_id=$1 for update`
	} else {
		query = match_select +
			`where ma.match_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&ma.MatchId, &ma.Status, &ma.StartTime, &ma.Players, &ma.Result, &ma.PlaceId, &ma.Sport, &ma.Teams, &ma.CreatedAt, &ma.CreatedBy, &ma.UpdatedAt, &ma.UpdatedBy, &ma.DeletedAt, &ma.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("match does not exist for username: %v", id)
		}
	}

	return ma, err
}

func (r *MatchCrud) GetCount(ctx context.Context, sp DR.MatchSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("MatchCrud.GetCount", L.Any("match", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := match_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("MatchCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *MatchCrud) Search(ctx context.Context, sp DR.MatchSearchParams, qa QueryAble) ([]DR.Match, error) {
	L.L.WithRequestID(ctx).Info("MatchCrud.Search", L.Any("match", sp))

	db := r.GetTx(qa)

	results := []DR.Match{}
	var params []interface{}

	query := match_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("MatchCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ma := DR.Match{}
		err := rows.Scan(&ma.MatchId, &ma.Status, &ma.StartTime, &ma.Players, &ma.Result, &ma.PlaceId, &ma.Sport, &ma.Teams, &ma.CreatedAt, &ma.CreatedBy, &ma.UpdatedAt, &ma.UpdatedBy, &ma.DeletedAt, &ma.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, ma)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("MatchCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a match
func (r *MatchCrud) Update(ctx context.Context, up DR.MatchUpdateParams, qa QueryAble, by *string) (DR.Match, error) {
	L.L.WithRequestID(ctx).Info("MatchCrud.Update", L.Any("match", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("MatchCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Match{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Match{}, fmt.Errorf("no rows affected")
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
