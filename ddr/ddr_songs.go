package ddr

import (
	"fmt"
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

// SongIds retrieves all song ids
func SongIds(client *http.Client) ([]string, error) {
	const musicDataURI = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_data_single.html?offset={page}&filter=0&filtertype=0&sorttype=0"
	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index="

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
	const musicDataURI = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_data_single.html?offset={page}&filter=0&filtertype=0&sorttype=0"

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

	const baseDetail = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_detail.html?index="

	wg := new(sync.WaitGroup)
	wg.Add(len(songIds))

	errorCount := 0

	for _, id := range songIds {
		go func(songId string, songList *[]Song) {
			defer wg.Done()
			doc, err := util.GetPageContentAsGoQuery(client, baseDetail + songId)
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
					fmt.Println(html)
				} else {

					imgPath, exists := img.First().Attr("src")
					if exists {
						fmt.Println(imgPath)
					}
					song.Image = imgPath
				}
			})

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