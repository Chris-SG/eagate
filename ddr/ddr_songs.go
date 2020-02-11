package ddr

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/eagate/util"
)

type Song struct {
	Id string `db:"song_id"`
	Name string `db:"song_name"`
	Artist string `db:"song_artist"`
	Image string `db:"song_image"`
}

var (
	mtx = &sync.Mutex{}
	lst = make([]string, 0)
)

func UpdateSongs(client *http.Client) error {
	songs, err := SongIds(client)
	if err != nil {
		return err
	}
	newSongs := GetSongIdsNotInDb(songs)
	songData, err := SongData(client, newSongs)
	if err != nil {
		return err
	}
	err = UpdateSongsDb(songData)
	return err
}

// SongIds retrieves all song ids
func SongIds(client *http.Client) ([]string, error) {
	const musicDataResource = "/game/ddr/ddra20/p/playdata/music_data_single.html?offset={page}&filter=0&filtertype=0&sorttype=0"
	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index="

	musicDataURI := util.BuildEaURI(musicDataResource)

	totalPages, err := songPageCount(client)
	if err != nil {
		return nil, err
	}

	fmt.Println(totalPages)

	wg := new(sync.WaitGroup)
	wg.Add(totalPages)

	errorCount := 0

	for idx := 0; idx < totalPages; idx++ {
		go func(currPage int) {
			defer wg.Done()

			currentPageURI := strings.Replace(musicDataURI, "{page}", strconv.Itoa(currPage), -1)
			doc, err := util.GetPageContentAsGoQuery(client, currentPageURI)

			if err != nil {
				errorCount++
				return
			}

			internalList := make([]string, 0)

			doc.Find("tr.data").Each(func(i int, s *goquery.Selection) {
				aElement := s.Find("a").First()
				href, exists := aElement.Attr("href")
				if exists {
					id := strings.Replace(href, baseDetail, "", -1)
					internalList = append(internalList, id)
				}
			})
			mtx.Lock()
			defer mtx.Unlock()
			lst = append(lst, internalList...)
		}(idx)
	}

	wg.Wait()

	if errorCount > 0 {
		return lst, fmt.Errorf("failed to load all songs ids, %d/%d pages failed", errorCount, totalPages)
	}

	return lst, nil
}

func songPageCount(client *http.Client) (int, error) {
	const musicDataResource = "/game/ddr/ddra20/p/playdata/music_data_single.html?offset={page}&filter=0&filtertype=0&sorttype=0"

	musicDataURI := util.BuildEaURI(musicDataResource)

	currentPageURI := strings.Replace(musicDataURI, "{page}", strconv.Itoa(0), -1)
	doc, err := util.GetPageContentAsGoQuery(client, currentPageURI)
	if err != nil {
		return 0, fmt.Errorf("failed to get music data page")
	}
	return doc.Find("div#paging_box").First().Find("div.page_num").Length(), nil
}

func SongData(client *http.Client, songIds []string) ([]Song, error) {

	var (
		songMtx = &sync.Mutex{}
		songLst = make([]Song, 0)
	)

	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index="

	baseDetailURI := util.BuildEaURI(baseDetail)

	wg := new(sync.WaitGroup)
	wg.Add(len(songIds))

	errorCount := 0

	for _, id := range songIds {
		go func(songId string, songList *[]Song) {
			defer wg.Done()
			doc, err := util.GetPageContentAsGoQuery(client, baseDetailURI + songId)
			fmt.Printf("Starting song id %s\n", songId)

			if err != nil {
				errorCount++
				return
			}

			song := Song{ Id: songId }

			doc.Find("table#music_info").First().Find("td").Each(func(i int, s *goquery.Selection) {
				img := s.Find("img")
				if img.Length() == 0 {
					html, _ := s.Html()
					songDataPair := strings.Split(html, "<br/>")
					song.Name = songDataPair[0]
					song.Artist = songDataPair[1]
				} else {

					imgPath, exists := img.First().Attr("src")
					if exists {
						imgUrl := fmt.Sprintf("https://p.eagate.573.jp%s", imgPath)
						imgData, err := client.Get(imgUrl)
						if err != nil {
							fmt.Println(err)
							errorCount++
						} else {
							body, err := ioutil.ReadAll(imgData.Body)
							if err == nil {
								song.Image = base64.StdEncoding.EncodeToString(body)
							}
						}
						imgData.Body.Close()
					}
				}
			})

			fmt.Printf("%s - %s - %s\n", song.Id, song.Name, song.Artist)

			songMtx.Lock()
			defer songMtx.Unlock()
			*songList = append(*songList, song)
		}(id, &songLst)
	}

	wg.Wait()

	if errorCount > 0 {
		return songLst, fmt.Errorf("failed to load all song data, %d/%d songs failed", errorCount, len(songIds))
	}

	return songLst, nil
}