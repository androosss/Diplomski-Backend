// Package crud(repo) contains code that is used for CRUD database operations in TRI PAY.
package crud

import (
	"context"
	"database/sql"
)

type Crud struct {
	db       *sql.DB
	crudRepo *Repo
}

func (c *Crud) SetCrudRepo(cr *Repo) {
	c.crudRepo = cr
}

func isQaNil(qa QueryAble) bool {
	switch v := qa.(type) {
	case (*sql.Tx):
		return v == nil
	case (*sql.DB):
		return v == nil
	}
	return true
}

func (c *Crud) GetTx(qa QueryAble) QueryAble {
	if !isQaNil(qa) {
		return qa
	}
	return c.db
}

type Creator interface {
	Create(ctx context.Context, entity interface{}, qa QueryAble, by *string) (interface{}, error)
}

type Updater interface {
	Update(ctx context.Context, entity interface{}, qa QueryAble, by *string) (interface{}, error)
}

type Deleter interface {
	Update(ctx context.Context, entity interface{}, qa QueryAble, by *string) error
}

type Searcher interface {
	Search(ctx context.Context, params interface{}, qa QueryAble) ([]interface{}, error)
}

type Exister interface {
	exists(ctx context.Context, entity interface{}, qa QueryAble) (bool, error)
}

type ByIdGetter interface {
	GetById(ctx context.Context, id int64, qa QueryAble) (interface{}, error)
}

type AllGetter interface {
	GetAll(ctx context.Context, qa QueryAble) ([]interface{}, error)
}

// https://github.com/golang/go/issues/14468
type QueryAble interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
