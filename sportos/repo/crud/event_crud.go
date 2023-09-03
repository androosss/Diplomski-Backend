package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type EventCrud struct {
	Crud
}

func InitEventCrud(db *sql.DB) *EventCrud {
	return &EventCrud{
		Crud{
			db: db,
		},
	}
}

const (
	event_select = `
		select ev.event_id, ev.name, ev.owner_id, ev.sport, ev.status, ev.time, ev.teams, ev.tournament, ev.created_at, ev.created_by, ev.updated_at, ev.updated_by, ev.deleted_at, ev.deleted_by
		from event ev
	`
	event_count = `select count(*) from event ev `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *EventCrud) checkConstraints(ctx context.Context, ev DR.Event, qa QueryAble) bool {
	L.L.WithRequestID(ctx).Info("EventCrud.checkConstraints", L.Any("id", ev))

	db := r.GetTx(qa)

	var count int
	row := db.QueryRowContext(ctx, `select count(*) from event where owner_id=$1 and name=$2`, ev.Owner, ev.Name)

	err := row.Scan(&count)
	if err != nil {
		L.L.Error("EventCrud.checkConstraints error", L.Any("err", err))
	}

	return count > 0
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a Event
func (r *EventCrud) Create(ctx context.Context, en DR.Event, qa QueryAble, by *string) (DR.Event, error) {
	L.L.WithRequestID(ctx).Info("EventCrud.Create", L.Any("event", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	if r.checkConstraints(ctx, en, qa) {
		return en, fmt.Errorf("event with name %s already exists for place %s", en.Name, en.Owner)
	}

	query := `insert into event (name, owner_id, sport, status, time, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6, $7) RETURNING event_id;`
	params := []interface{}{en.Name, en.Owner, en.Sport, en.Status, en.Time, en.CreatedAt, en.CreatedBy}

	L.L.Debug("EventCrud.Create insert", L.String("query", query), L.Any("params", params))

	err := db.QueryRowContext(ctx, query, params...).Scan(&en.EventId)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}
	pen, err := r.GetById(ctx, en.EventId, qa)
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

// GetById returns event by userrname
func (r *EventCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.Event, error) {
	L.L.WithRequestID(ctx).Info("EventCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	ev := DR.Event{}
	query := ""
	if qa != nil {
		query = event_select +
			`where ev.event_id=$1 for update`
	} else {
		query = event_select +
			`where ev.event_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&ev.EventId, &ev.Name, &ev.Owner, &ev.Sport, &ev.Status, &ev.Time, &ev.Teams, &ev.Tournament, &ev.CreatedAt, &ev.CreatedBy, &ev.UpdatedAt, &ev.UpdatedBy, &ev.DeletedAt, &ev.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("event does not exist for username: %v", id)
		}
	}

	return ev, err
}

func (r *EventCrud) GetCount(ctx context.Context, sp DR.EventSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("EventCrud.GetCount", L.Any("event", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := event_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("EventCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *EventCrud) Search(ctx context.Context, sp DR.EventSearchParams, qa QueryAble) ([]DR.Event, error) {
	L.L.WithRequestID(ctx).Info("EventCrud.Search", L.Any("event", sp))

	db := r.GetTx(qa)

	results := []DR.Event{}
	var params []interface{}

	query := event_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("EventCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ev := DR.Event{}
		err := rows.Scan(&ev.EventId, &ev.Name, &ev.Owner, &ev.Sport, &ev.Status, &ev.Time, &ev.Teams, &ev.Tournament, &ev.CreatedAt, &ev.CreatedBy, &ev.UpdatedAt, &ev.UpdatedBy, &ev.DeletedAt, &ev.DeletedBy)
		if err != nil {
			return nil, err
		}
		results = append(results, ev)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("EventCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a event
func (r *EventCrud) Update(ctx context.Context, up DR.EventUpdateParams, qa QueryAble, by *string) (DR.Event, error) {
	L.L.WithRequestID(ctx).Info("EventCrud.Update", L.Any("event", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("EventCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.Event{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.Event{}, fmt.Errorf("no rows affected")
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
