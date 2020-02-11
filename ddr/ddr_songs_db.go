package ddr

import (
	"github.com/chris-sg/eagate/ea_db"
	"github.com/jmoiron/sqlx"
)

func LoadSongsDB(db *sqlx.DB) ([]Song, error) {
	queryString := `SELECT * FROM "ddrSongs"`
	rows, err := ea_db.ExecuteQuery(db, queryString)
	if err != nil {
		return nil, err
	}

	var songs []Song
	for rows.Next() {
		var song Song
		err = rows.StructScan(&song)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}