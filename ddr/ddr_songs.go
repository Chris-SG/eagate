package ddr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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

	totalPages := 0
	songsRead := 0

	{
		currentPageURI := strings.Replace(musicDataURI, "{page}", strconv.Itoa(0), -1)
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
				body = util.ShiftJISBytesToUTF8Bytes(body)
			}
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))

		if err != nil {
			log.Fatal(err)
		}

		totalPages = doc.Find("div#paging_box").First().Find("div.page_num").Length()
	}

	fmt.Println(totalPages)

	wg := new(sync.WaitGroup)
	wg.Add(totalPages)

	for idx := 0; idx < totalPages; idx++ {
		go func(currPage int) {
			defer wg.Done()
			fmt.Printf("working on page %d\n", currPage)

			currentPageURI := strings.Replace(musicDataURI, "{page}", strconv.Itoa(idx), -1)
			res, err := client.Get(currentPageURI)

			if err != nil {
				log.Fatal(err)
			}

			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)

			contentType, ok := res.Header["Content-Type"]
			if ok && len(contentType) > 0 {
				if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
					body = util.ShiftJISBytesToUTF8Bytes(body)
				}
			}

			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))

			if err != nil {
				log.Fatal(err)
			}

			internalList := make([]string, 0)

			doc.Find("tr.data").Each(func(i int, s *goquery.Selection) {
				fmt.Println("Found one...")
				aElement := s.Find("a").First()
				href, exists := aElement.Attr("href")
				if exists {
					id := strings.Replace(href, baseDetail, "", -1)
					internalList = append(internalList, id)
					songsRead++
				}
			})
			defer mtx.Unlock()
			mtx.Lock()
			lst = append(lst, internalList...)
		}(idx)
	}

	return lst, nil
}

func SongData(client *http.Client, songIds []string) []Song {
	const baseDetail = "/game/ddr/ddra20/p/playdata/music_detail.html?index="


	for _, id := range songIds {
		res, err := client.Get(baseDetail + id)
		defer res.Body.Close()

		if err != nil {
			fmt.Println(err)
			continue
		}

		body, err := ioutil.ReadAll(res.Body)

		contentType, ok := res.Header["Content-Type"]
		if ok && len(contentType) > 0 {
			if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
				body = util.ShiftJISBytesToUTF8Bytes(body)
			}
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))

		if err != nil {
			fmt.Println(err)
			continue
		}

		doc.Find("table#music_info").First().Find("td").Each(func(i int, s *goquery.Selection) {
			img := s.Find("img")
			if img.Length() == 0 {
				fmt.Println(s.Contents().Text())
			}
		})
	}
	return make([]Song, 0)
}