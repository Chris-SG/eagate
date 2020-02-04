package ddr

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/chris-sg/eagate/util"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	//"github.com/chris-sg/eagate/util"
	"github.com/PuerkitoBio/goquery"
)

type Song struct {
	Id string `db:"song_id"`
	Name string `db:"song_name"`
	Artist string `db:"song_artist"`
	Image string `db:"song_image"`
}

var (
	mtx = &sync.Mutex{}
	lst = list.New()
)

// SongIds retrieves all song ids
func SongIds(client *http.Client) (*list.List, error) {
	const musicDataURI = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_data_single.html?offset={page}&filter=0&filtertype=0&sorttype=0"
	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index="
	const maxSongsPerPage = 50

	totalPages := 0
	songsRead := maxSongsPerPage

	{
		currentPageURI := strings.Replace(musicDataURI, "{page}", strconv.Itoa(page), -1)
		res, err := client.Get(currentPageURI)

		if err != nil {
			fmt.Print(err)
			return lst, err
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		contentType, ok := res.Header["Content-Type"]
		if ok && len(contentType) > 0 {
			if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
				//body = util.ShiftJISBytesToUTF8Bytes(body)
			}
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))

		if err != nil {
			log.Fatal(err)
		}

		doc.Find("div#paging_box").First().Find("div.page_num").Each(func(i int, s *goquery.Selection) {
			totalPages++
		})
	}



	return lst, nil
}