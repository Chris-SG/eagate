package ddr

import (
	"database/sql"
	"github.com/chris-sg/eagate/ea_db"
)

func LoadSongsDB(db *sql.DB) ([]Song, error) {
	queryString := `SELECT * FROM "ddrSongs"`
	rows, err := ea_db.ExecuteQuery(db, queryString)
	if err != nil {
		return nil, err
	}

	var songs []Song
	err = rows.Scan(&songs)
	if err != nil {
		return nil, err
	}

	return songs, nil
}