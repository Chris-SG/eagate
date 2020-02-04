package db

import (
	"database/sql"
	"fmt"
	"log"
)

func createDbConnection(user string, password string, dbname string) *sql.DB {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s", user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func executeQuery(db *sql.DB, query string) (sql.Result, error) {
	result, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	return result, nil
}