package crud

import (
	"backend/internal/cache"
	L "backend/internal/logging"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
	"time"
)

// DBConnection contains parameters for creating the database connection
type DBConnection struct {
	DBName   string
	Host     string
	Port     string
	User     string
	Password string
}

type Repo struct {
	DB             *sql.DB
	PlayerCrud     *PlayerCrud
	CoachCrud      *CoachCrud
	PlaceCrud      *PlaceCrud
	EventCrud      *EventCrud
	ApiJournalCrud *ApiJournalCrud
	AuditCrud      *AuditCrud
	UserCrud       *UserCrud
	MatchCrud      *MatchCrud
	PracticeCrud   *PracticeCrud
	TeamCrud       *TeamCrud
	UserPostsCrud  *UserPostCrud
	NameCache      *cache.Cache[string, string]
}

func (dbCon *DBConnection) InitRepo() *Repo {
	connectionString := fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=disable", dbCon.DBName, dbCon.Host, dbCon.Port, dbCon.User, dbCon.Password)

	postgreDb, err := sql.Open("postgres", connectionString)
	if err != nil {
		L.L.Fatal("Could not instantiate database")
	}
	err = postgreDb.Ping()
	if err != nil {
		L.L.Fatal("Could not connect to database")
	}
	postgreDb.SetMaxOpenConns(100)
	postgreDb.SetMaxIdleConns(100)
	postgreDb.SetConnMaxLifetime(5 * time.Minute)

	r := &Repo{
		DB:             postgreDb,
		PlayerCrud:     InitPlayerCrud(postgreDb),
		CoachCrud:      InitCoachCrud(postgreDb),
		PlaceCrud:      InitPlaceCrud(postgreDb),
		EventCrud:      InitEventCrud(postgreDb),
		ApiJournalCrud: InitApiJournalCrud(postgreDb),
		AuditCrud:      InitAuditCrud(postgreDb),
		UserCrud:       InitUserCrud(postgreDb),
		TeamCrud:       InitTeamCrud(postgreDb),
		MatchCrud:      InitMatchCrud(postgreDb),
		PracticeCrud:   InitPracticeCrud(postgreDb),
		UserPostsCrud:  InitUserPostCrud(postgreDb),
	}
	r.PlayerCrud.SetCrudRepo(r)
	r.CoachCrud.SetCrudRepo(r)
	r.PlaceCrud.SetCrudRepo(r)
	r.EventCrud.SetCrudRepo(r)
	r.ApiJournalCrud.SetCrudRepo(r)
	r.AuditCrud.SetCrudRepo(r)
	r.UserCrud.SetCrudRepo(r)
	r.MatchCrud.SetCrudRepo(r)
	r.PracticeCrud.SetCrudRepo(r)
	r.TeamCrud.SetCrudRepo(r)
	r.UserPostsCrud.SetCrudRepo(r)

	r.NameCache = cache.NewCache[string, string]()
	return r
}

func (r *Repo) GetImageName(ctx context.Context) (string, error) {
	var seq string
	err := r.DB.QueryRowContext(ctx, "select nextval('image_seq');").Scan(&seq)
	hasher := sha256.New()
	hasher.Write([]byte(seq))
	return base32.StdEncoding.EncodeToString(hasher.Sum(nil)), err
}
