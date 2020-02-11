package ea_db

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

func CreateDbConnection(user string, password string, dbname string, host string) (*sqlx.DB, error) {
	connStr := fmt.Sprintf( "user=%s password=%s dbname=%s host=%s sslmode=disable", user, password, dbname, host)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ExecuteQuery(db *sqlx.DB, query string) (*sqlx.Rows, error) {
	rows, err := db.Queryx(query)
	if err != nil {
		log.Fatal(err)
	}
	return rows, nil
}