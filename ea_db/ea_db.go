package ea_db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func CreateDbConnection(user string, password string, dbname string, host string) (*sql.DB, error) {
	connStr := fmt.Sprintf( "user=%s password=%s dbname=%s host=%s sslmode=disable", user, password, dbname, host)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ExecuteQuery(db *sql.DB, query string) (*sql.Rows, error) {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	return rows, nil
}