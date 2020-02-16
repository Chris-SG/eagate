package ddr



//////////////////////
// Score Info Block //
//////////////////////

// GetScoreInfo will process the provided song ids and retrieve all
// player score information for these songs.
/*func GetScoreInfo(client *http.Client, charts []ddr_models.SongDifficulty) ([]ddr_models.SongStatistics, error) {
	const musicDetail = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_detail.html?index={id}&diff={diff}"

	var songStatistics []ddr_models.SongStatistics
	wg := new(sync.WaitGroup)

	for _, chart := range charts {
			wg.Add(1)
			go func(songId string, currDiff int8) {
				defer wg.Done()
				score, err := LoadChartDifficultyInfo(client, songId, currDiff)
				if err != nil {
					fmt.Println(err)
				}
				score.Level = difficulties[currDiff]
				reflect.ValueOf(&chart).Elem().FieldByIndex([]int{currDiff}).Set(reflect.ValueOf(score))
			}(chart.SongId, chart.DifficultyId)

		wg.Wait()

		scoreInfo[element] = chart
	}

	printScores(scoreInfo)
	return nil
}

/////////////////////////
// Recent Scores Block //
/////////////////////////

// RecentScore defines a recent score listing.
type RecentScore struct {
	id         string
	difficulty int8
	score      int
	time       time.Time
	cleared    bool
}

// RecentScores will load all recent (up top last 50) songs
// and scores.
func RecentScores(client *http.Client) ([]RecentScore, error) {
	const recentSongs = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_recent.html"
	const maxRecentScores = 50
	recentScores := make([]RecentScore, maxRecentScores)

	res, err := client.Get(recentSongs)
	if err != nil {
		fmt.Print(err)
		return recentScores, err
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
		fmt.Print(err)
		return recentScores, err
	}

	table := doc.Find("table#data_tbl").First()
	if table.Length() == 0 {
		return recentScores, fmt.Errorf("Could not find data_tbl")
	}

	table.Find("a.music_info.cboxelement").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if exists {
			difficulty, err := strconv.Atoi(href[len(href)-1:])
			if err == nil {
				recentScores[i].difficulty = int8(difficulty)
			}
			recentScores[i].id = href[strings.Index(href, "=")+1 : strings.Index(href, "&")]
		}
		return i != maxRecentScores
	})

	table.Find("td.rank").Each(func(i int, s *goquery.Selection) {
		img := s.Find("img").First()
		path, exists := img.Attr("src")
		if exists {
			recentScores[i].cleared = !strings.Contains(path, "rank_s_e")
		}
	})

	table.Find("td.score").Each(func(i int, s *goquery.Selection) {
		recentScores[i].score, _ = strconv.Atoi(s.Text())
	})

	format := "2006-01-02 15:04:05"
	table.Find("td.date").Each(func(i int, s *goquery.Selection) {
		t, err := time.Parse(format, s.Text())
		if err == nil {
			recentScores[i].time = t
		}
	})

	return recentScores, nil
}
*/