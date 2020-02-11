package ddr

import (
	"fmt"
	"github.com/chris-sg/eagate/ea_db"
)

func LoadSongIdsDB() ([]string, error) {
	ids := []string{}
	db := ea_db.GetDb()

	err := db.Get(&ids, `SELECT song_id FROM "ddrSongs"`)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func LoadSongsDB() ([]Song, error) {
	queryString := `SELECT * FROM "ddrSongs"`
	rows, err := ea_db.ExecuteQuery(queryString)
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

func UpdateSongsDb(songs []Song) error {
	db := ea_db.GetDb()

	result, err := db.NamedExec(`INSERT INTO "ddrSongs" ('song_id', 'song_name', 'song_artist', 'song_image') VALUES (:Id, :Name, :Artist, :Image) ON CONFLICT ('song_id') DO NOTHING`, songs)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("UpdateSongsDb: %d rows affected", rowsAffected)
	return nil
}

func GetSongIdsNotInDb(ids []string) []string {
	var dbIds []string
	rows, err := ea_db.ExecuteQuery(`SELECT user_id FROM "ddrSongs"`)
	if err != nil {
		return nil
	}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		dbIds = append(dbIds, id)
	}

	check := map[string]struct{}{}
	result := []string{}

	for _, id := range dbIds {
		check[id] = struct{}{}
	}

	for _, id := range ids {
		if _, ok := check[id]; !ok {
			result = append(result, id)
		}
	}

	return result
}