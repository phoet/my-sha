package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
	"log"
	"time"
)

func (i *Repo) PreInsert(s gorp.SqlExecutor) error {
	i.Created = time.Now()
	i.Updated = i.Created
	return nil
}

func (i *Repo) PreUpdate(s gorp.SqlExecutor) error {
	i.Updated = time.Now()
	return nil
}

type Repo struct {
	Id       int64 `db:"id"`
	App      string
	Revision string
	Token    string
	Created  time.Time
	Updated  time.Time
}

func newRepo(app string) Repo {
	return Repo{
		App:   app,
		Token: generateUUID(),
	}
}

func initDb() *gorp.DbMap {
	url := mustGetenv("DATABASE_URL")
	db, err := sql.Open("postgres", url)
	checkErr(err, "sql.Open failed")

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}

	dbmap.AddTableWithName(Repo{}, "repos").SetKeys(true, "Id")

	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Panicln(msg, err)
	}
}
