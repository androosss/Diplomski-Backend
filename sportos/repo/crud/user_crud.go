package crud

import (
	L "backend/internal/logging"
	DR "backend/sportos/repo/dto"
	"backend/sportos/repo/util"
	"context"
	"database/sql"
	"fmt"
)

type UserCrud struct {
	Crud
}

func InitUserCrud(db *sql.DB) *UserCrud {
	return &UserCrud{
		Crud{
			db: db,
		},
	}
}

const (
	user_select = `
		select usr.user_id, usr.email, usr.email_verified, usr.user_type, usr.password_hash, usr.token, usr.token_valid_until, usr.token_refresh_until, usr.created_at, usr.created_by, usr.updated_at, usr.updated_by, usr.deleted_at, usr.deleted_by
		from "user" usr
	`
	user_count = `select count(*) from "user" usr `
)

////////////////////////////////////////////////UTIL/////////////////////////////////////////////////////////////////////////////////////

func (r *UserCrud) exists(ctx context.Context, id string, qa QueryAble) bool {
	L.L.WithRequestID(ctx).Info("UserCrud.exists", L.Any("id", id))

	db := r.GetTx(qa)

	var count int
	row := db.QueryRowContext(ctx, `select count(*) from "user" where user_id=$1`, id)

	err := row.Scan(&count)
	if err != nil {
		L.L.Error("UserCrud.exists error", L.Any("err", err))
	}

	return count > 0
}

func trimMail(s string) string {
	if len(s) > 7 && s[0:7] == "google_" {
		return s[7:]
	}
	if len(s) > 9 && s[0:9] == "facebook_" {
		return s[9:]
	}
	return s
}

////////////////////////////////////////////////CREATE///////////////////////////////////////////////////////////////////////////////////

// Creates a user
func (r *UserCrud) Create(ctx context.Context, en DR.User, qa QueryAble, by *string) (DR.User, error) {
	L.L.WithRequestID(ctx).Info("UserCrud.Create", L.Any("user", en))

	db := r.GetTx(qa)

	if en.CreatedAt.IsZero() {
		en.EditInfoC = DR.CreateEditInfoC(by)
	}

	if r.exists(ctx, en.Username, qa) {
		return en, fmt.Errorf("user with username %s already exists", en.Username)
	}

	query := `insert into "user" (user_id, email, email_verified, user_type, password_hash, created_at, created_by)
	values ($1, $2, $3, $4, $5, $6, $7) RETURNING user_id;`
	params := []interface{}{en.Username, en.Email, en.EmailVerified, en.UserType, en.PasswordHash, en.CreatedAt, en.CreatedBy}

	L.L.Debug("UserCrud.Create insert", L.String("query", query), L.Any("params", params))

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

// GetById returns user by username
func (r *UserCrud) GetById(ctx context.Context, id string, qa QueryAble) (DR.User, error) {
	L.L.WithRequestID(ctx).Info("UserCrud.GetById", L.String("username", id))

	db := r.GetTx(qa)

	usr := DR.User{}
	query := ""
	if qa != nil {
		query = user_select +
			`where usr.user_id=$1 for update`
	} else {
		query = user_select +
			`where usr.user_id=$1`
	}
	row := db.QueryRowContext(ctx, query,
		id)

	err := row.Scan(&usr.Username, &usr.Email, &usr.EmailVerified, &usr.UserType, &usr.PasswordHash, &usr.Token, &usr.TokenValidUntil, &usr.TokenRefreshUntil, &usr.CreatedAt, &usr.CreatedBy, &usr.UpdatedAt, &usr.UpdatedBy, &usr.DeletedAt, &usr.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("user does not exist for username: %v", id)
		}
	}
	usr.Email = trimMail(usr.Email)

	return usr, err
}

// GetById returns user by email
func (r *UserCrud) GetByEmail(ctx context.Context, email string, qa QueryAble) (DR.User, error) {
	L.L.WithRequestID(ctx).Info("UserCrud.GetById", L.String("email", email))

	db := r.GetTx(qa)

	usr := DR.User{}
	query := ""
	if qa != nil {
		query = user_select +
			`where usr.email=$1 for update`
	} else {
		query = user_select +
			`where usr.email=$1`
	}
	row := db.QueryRowContext(ctx, query,
		email)

	err := row.Scan(&usr.Username, &usr.Email, &usr.EmailVerified, &usr.UserType, &usr.PasswordHash, &usr.Token, &usr.TokenValidUntil, &usr.TokenRefreshUntil, &usr.CreatedAt, &usr.CreatedBy, &usr.UpdatedAt, &usr.UpdatedBy, &usr.DeletedAt, &usr.DeletedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("user does not exist for email: %v", email)
		}
	}
	usr.Email = trimMail(usr.Email)

	return usr, err
}

func (r *UserCrud) GetCount(ctx context.Context, sp DR.UserSearchParams, qa QueryAble) (int, error) {
	L.L.WithRequestID(ctx).Info("UserCrud.GetCount", L.Any("user", sp))

	db := r.GetTx(qa)

	var params []interface{}

	query := user_count

	err := DR.AppendCountQuery(&sp, &query, &params)
	if err != nil {
		return 0, err
	}

	L.L.WithRequestID(ctx).Debug("UserCrud.GetCount query", L.Any("query", L.String("query", query)))

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

func (r *UserCrud) Search(ctx context.Context, sp DR.UserSearchParams, qa QueryAble) ([]DR.User, error) {
	L.L.WithRequestID(ctx).Info("UserCrud.Search", L.Any("user", sp))

	db := r.GetTx(qa)

	results := []DR.User{}
	var params []interface{}

	query := user_select

	err := DR.AppendQuery(&sp, &query, &params)
	if err != nil {
		return nil, err
	}

	L.L.WithRequestID(ctx).Debug("UserCrud.Search query", L.Any("query", L.String("query", query)))

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		usr := DR.User{}
		err := rows.Scan(&usr.Username, &usr.Email, &usr.EmailVerified, &usr.UserType, &usr.PasswordHash, &usr.Token, &usr.TokenValidUntil, &usr.TokenRefreshUntil, &usr.CreatedAt, &usr.CreatedBy, &usr.UpdatedAt, &usr.UpdatedBy, &usr.DeletedAt, &usr.DeletedBy)
		if err != nil {
			return nil, err
		}
		usr.Email = trimMail(usr.Email)
		results = append(results, usr)
	}

	if len(results) == 0 {
		L.L.WithRequestID(ctx).Warn("UserCrud.Search No rows returned ")
	}
	return results, nil
}

////////////////////////////////////////////////UPDATE///////////////////////////////////////////////////////////////////////////////////

// updates a user
func (r *UserCrud) Update(ctx context.Context, up DR.UserUpdateParams, qa QueryAble, by *string) (DR.User, error) {
	L.L.WithRequestID(ctx).Info("UserCrud.Update", L.Any("user", up))

	up.PopulateUpdateFields(by)

	old, _ := r.GetById(ctx, up.Id, qa)

	db := r.GetTx(qa)
	var query string
	params := []interface{}{}

	DR.AppendUpdateQuery(up, &query, &params)

	L.L.Debug("UserCrud.Update update", L.String("query", query), L.Any("params", params))

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		util.LogPqError(ctx, err)
		return DR.User{}, err
	}

	ra, _ := result.RowsAffected()
	if ra == 0 {
		return DR.User{}, fmt.Errorf("no rows affected")
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
