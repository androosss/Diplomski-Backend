package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type TeamCrud struct {
	Crud
}

func InitTeamCrud(db *sql.DB) *TeamCrud {
	return &TeamCrud{
		Crud{
			db: db,
		},
	}
}

const (
	team_select = `
		select te.team_id, te.name, te.sport, te.status, te.players, te.created_at, te.created_by, te.updated_at, te.updated_by, te.deleted_at, te.deleted_by
		from team te
	`
	team_count = `select count(*) from team te `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *TeamCrud) CheckConstraints(ctx context.Context, te DR.Team, qa QueryAble) bool {
	L.L.WithRequestID(ctx).Info("TeamCrud.checkConstraints", L.Any("id", te))

	db := r.GetTx(qa)

	var count int
	row := db.QueryRowContext(ctx, `select count(*) from team where sport=$1 and name=$2`, te.Sport, te.Name)

	err := row.Scan(&count)
	if err != nil {
		L.L.Error("TeamCrud.checkConstraints error", L.Any("err", err))
	}

	return count > 0
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a Team
func (r *TeamCrud) Create(ctx context.Context, en DR.Team, qa QueryAble, by *string) (DR.Team, error) {
	L.L.WithRequestID(ctx).Info("TeamCrud.Create", L.Any("team", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	if r.CheckConstraints(ctx, en, qa) {
		return en, fmt.Errorf("team with name %s already exists for sport %s", en.Name, en.Sport)
	}

	query := `insert into team (name, sport, status, players, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6) RETURNING team_id;`
	params := []interface{}{en.Name, en.Sport, en.Status, en.Players, en.CreatedAt, en.CreatedBy}

	L.L.Debug("TeamCrud.Create insert", L.String("query", query), L.Any("params", params))

	err := db.QueryRowContext(ctx, query, params...).Scan(&en.TeamId)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.TeamId, qa)
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

// GetById returns team by id
func (r *TeamCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Team, error) {
	L.L.WithRequestID(ctx).Info("TeamCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	te := DR.Team{}
	query := ""
	if qa != nil {
		query = team_select +
			`where te.team_id=$1 for update`
	} else {
		query = team_select +
			`where te.team_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&te.TeamId, &te.Name, &te.Sport, &te.Status, &te.Players, &te.CreatedAt, &te.CreatedBy, &te.UpdatedAt, &te.UpdatedBy, &te.DeletedAt, &te.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("team does not exist for username: %v", id)
		}
	}

	return te, err
}

func (r *TeamCrud) GetCount(ctx context.Context, sp DR.TeamSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("TeamCrud.GetCount", L.Any("team", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := team_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("TeamCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *TeamCrud) Search(ctx context.Context, sp DR.TeamSearchParams, qa QueryAble) ([]DR.Team, error) {
	L.L.WithRequestID(ctx).Info("TeamCrud.Search", L.Any("team", sp))

	db := r.GetTx(qa)

	results := []DR.Team{}
	var params []interface{}

	query := team_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("TeamCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		te := DR.Team{}
		err := rows.Scan(&te.TeamId, &te.Name, &te.Sport, &te.Status, &te.Players, &te.CreatedAt, &te.CreatedBy, &te.UpdatedAt, &te.UpdatedBy, &te.DeletedAt, &te.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, te)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("TeamCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a team
func (r *TeamCrud) Update(ctx context.Context, up DR.TeamUpdateParams, qa QueryAble, by *string) (DR.Team, error) {
	L.L.WithRequestID(ctx).Info("TeamCrud.Update", L.Any("team", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("TeamCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Team{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Team{}, fmt.Errorf("no rows affected")
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
