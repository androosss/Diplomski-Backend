package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type UserPostCrud struct {
	Crud
}

func InitUserPostCrud(db *sql.DB) *UserPostCrud {
	return &UserPostCrud{
		Crud{
			db: db,
		},
	}
}

const (
	userpost_select = `
		select up.user_id, up.user_text, up.image_names, up.created_at
		from userpost up
	`
	userpost_count = `select count(*) from userpost up `
)

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a UserPost
func (r *UserPostCrud) Create(ctx context.Context, en DR.UserPost, qa QueryAble, by *string) (DR.UserPost, error) {
	L.L.WithRequestID(ctx).Info("UserPostCrud.Create", L.Any("userpost", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	query := `insert into userpost (user_id, user_text, image_names, created_at)
	values ($1, $2, $3, $4);`
	params := []interface{}{en.UserId, en.UserText, pq.Array(en.ImageNames), en.CreatedAt}

	L.L.Debug("UserPostCrud.Create insert", L.String("query", query), L.Any("params", params))

	_, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return en, err
	}

	return en, nil
}

////////////////////////////////////////////////READ/////////////////////////////////////////////////////////////////////////////////////

func (r *UserPostCrud) GetCount(ctx context.Context, sp DR.UserPostSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("UserPostCrud.GetCount", L.Any("userpost", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := userpost_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("UserPostCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *UserPostCrud) Search(ctx context.Context, sp DR.UserPostSearchParams, qa QueryAble) ([]DR.UserPost, error) {
	L.L.WithRequestID(ctx).Info("UserPostCrud.Search", L.Any("userpost", sp))

	db := r.GetTx(qa)

	results := []DR.UserPost{}
	var params []interface{}

	query := userpost_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("UserPostCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		up := DR.UserPost{}
		err := rows.Scan(&up.UserId, &up.UserText, pq.Array(&up.ImageNames), &up.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, up)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("UserPostCrud.Search No rows returned ")
	}
	return results, nil
}
