package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type PlayerCrud struct {
	Crud
}

func InitPlayerCrud(db *sql.DB) *PlayerCrud {
	return &PlayerCrud{
		Crud{
			db: db,
		},
	}
}

const (
	player_select = `
		select pl.user_id, pl.name, pl.city, pl.preferences, pl.statistics, pl.created_at, pl.created_by, pl.updated_at, pl.updated_by, pl.deleted_at, pl.deleted_by
		from player pl
	`
	player_count = `select count(*) from player pl `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *PlayerCrud) exists(ctx context.Context, id string, qa QueryAble) bool {
	L.L.WithRequestID(ctx).Info("PlayerCrud.exists", L.Any("id", id))

	db := r.GetTx(qa)

	var count int
	row := db.QueryRowContext(ctx, `select count(*) from player where user_id=$1`, id)

	err := row.Scan(&count)
	if err != nil {
		L.L.Error("PlayerCrud.exists error", L.Any("err", err))
	}

	return count > 0
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a player
func (r *PlayerCrud) Create(ctx context.Context, en DR.Player, qa QueryAble, by *string) (DR.Player, error) {
	L.L.WithRequestID(ctx).Info("PlayerCrud.Create", L.Any("player", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	if r.exists(ctx, en.Username, qa) {
		return en, fmt.Errorf("user with username %s already exists", en.Username)
	}

	query := `insert into player (user_id, name, city, created_at, created_by)
	values ($1, $2, $3, $4, $5) RETURNING user_id;`
	params := []interface{}{en.Username, en.Name, en.City, en.CreatedAt, en.CreatedBy}

	L.L.Debug("PlayerCrud.Create insert", L.String("query", query), L.Any("params", params))

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

// GetById returns player by userrname
func (r *PlayerCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Player, error) {
	L.L.WithRequestID(ctx).Info("PlayerCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	pl := DR.Player{}
	query := ""
	if qa != nil {
		query = player_select +
			`where pl.user_id=$1 for update`
	} else {
		query = player_select +
			`where pl.user_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&pl.Username, &pl.Name, &pl.City, &pl.Preferences, &pl.Statistics, &pl.CreatedAt, &pl.CreatedBy, &pl.UpdatedAt, &pl.UpdatedBy, &pl.DeletedAt, &pl.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("player does not exist for username: %v", id)
		}
	}

	return pl, err
}

// GetById returns player by email
func (r *PlayerCrud) GetByEmail(ctx context.Context, email string, qa QueryAble) (DR.Player, error) {
	L.L.WithRequestID(ctx).Info("PlayerCrud.GetById", L.String("email", email))

	db := r.GetTx(qa)

	pl := DR.Player{}
	query := ""
	if qa != nil {
		query = player_select +
			`where pl.email=$1 for update`
	} else {
		query = player_select +
			`where pl.email=$1`
	}
	row := db.QueryRowContext(ctx, query,
		email)

	err := row.Scan(&pl.Username, &pl.Name, &pl.City, &pl.Preferences, &pl.Statistics, &pl.CreatedAt, &pl.CreatedBy, &pl.UpdatedAt, &pl.UpdatedBy, &pl.DeletedAt, &pl.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("player does not exist for email: %v", email)
		}
	}

	return pl, err
}

func (r *PlayerCrud) GetCount(ctx context.Context, sp DR.PlayerSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("PlayerCrud.GetCount", L.Any("player", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := player_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("PlayerCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *PlayerCrud) Search(ctx context.Context, sp DR.PlayerSearchParams, qa QueryAble) ([]DR.Player, error) {
	L.L.WithRequestID(ctx).Info("PlayerCrud.Search", L.Any("player", sp))

	db := r.GetTx(qa)

	results := []DR.Player{}
	var params []interface{}

	query := player_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("PlayerCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		pl := DR.Player{}
		err := rows.Scan(&pl.Username, &pl.Name, &pl.City, &pl.Preferences, &pl.Statistics, &pl.CreatedAt, &pl.CreatedBy, &pl.UpdatedAt, &pl.UpdatedBy, &pl.DeletedAt, &pl.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, pl)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("PlayerCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a player
func (r *PlayerCrud) Update(ctx context.Context, up DR.PlayerUpdateParams, qa QueryAble, by *string) (DR.Player, error) {
	L.L.WithRequestID(ctx).Info("PlayerCrud.Update", L.Any("player", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("PlayerCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Player{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Player{}, fmt.Errorf("no rows affected")
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
