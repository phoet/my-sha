package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	// _ "github.com/mattn/go-sqlite3"
	"gopkg.in/gorp.v1"
	"log"
	"time"
)

func main() {
	dbmap := initDb()
	defer dbmap.Db.Close()

	err := dbmap.TruncateTables()
	checkErr(err, "TruncateTables failed")

	repo := newRepo("app123")

	err = dbmap.Insert(&repo)
	checkErr(err, "Insert failed")

	count, err := dbmap.SelectInt("select count(*) from repos")
	checkErr(err, "select count(*) failed")
	log.Println("Rows after inserting:", count)

	repo.Revision = "AABBCCDD"
	count, err = dbmap.Update(&repo)
	checkErr(err, "Update failed")
	log.Println("Rows updated:", count)

	err = dbmap.SelectOne(&repo, "select * from repos where id=$1", repo.Id)
	checkErr(err, "SelectOne failed")
	log.Println("repo row:", repo)

	var repos []Repo
	_, err = dbmap.Select(&repos, "select * from repos")
	checkErr(err, "Select failed")
	log.Println("All rows:")
	for x, r := range repos {
		log.Printf("    %d: %v\n", x, r)
	}

	log.Println("Done!")
}

type Repo struct {
	Id       int64 `db:"id"`
	App      string
	Revision string
	Token    string
	Created  int64
}

func newRepo(app string) Repo {
	return Repo{
		Created: time.Now().UnixNano(),
		App:     app,
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
		log.Fatalln(msg, err)
	}
}
