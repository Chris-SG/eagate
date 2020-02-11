package ea_db

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

var (
	activeDb *sqlx.DB
)

func CreateDbConnection(user string, password string, dbname string, host string) (error) {
	connStr := fmt.Sprintf( "user=%s password=%s dbname=%s host=%s sslmode=disable", user, password, dbname, host)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	activeDb = db

	return nil
}

func ExecuteQuery(query string) (*sqlx.Rows, error) {
	rows, err := activeDb.Queryx(query)
	if err != nil {
		log.Fatal(err)
	}
	return rows, nil
}

func GetDb() *sqlx.DB {
	return activeDb
}