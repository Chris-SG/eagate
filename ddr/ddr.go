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

// LoadChartDifficultyInfo will attempt to retrieve score information for a given difficulty.
// id must match a Song ID (from DDRMusicList) and difficulty must be from 0 to 8.
// Will return a Score on success, or an error.
func LoadChartDifficultyInfo(client *http.Client, id string, difficulty int8) (*ddr_models.SongStatistics, error) {
	const musicDetailResource = "/game/ddr/ddra20/p/playdata/music_detail.html?index={id}&diff={diff}"

	musicDiffDetailsResource := strings.Replace(musicDetailResource, "{id}", id, -1)
	musicDiffDetailsResource = strings.Replace(musicDiffDetailsResource, "{diff}", strconv.Itoa(int(difficulty)), -1)

	musicDiffDetailsUri := util.BuildEaURI(musicDiffDetailsResource)
	doc, err := util.GetPageContentAsGoQuery(client, musicDiffDetailsUri)

	if err != nil {
		return nil, err
	}

	songStatistics := ddr_models.SongStatistics{}

	scoreType := reflect.TypeOf(songStatistics)

	if !bytes.Contains(body, []byte("NO PLAY...")) {
		musicDetailTable := doc.Find("table#music_detail_table")
		pairs, err := util.TableThTd(musicDetailTable)

		if err != nil {
			fmt.Println(err)
			return score, err
		}

		val, found := pairs["フルコンボ種別"]
		if found {
			score.Lamp = ddrGetLampMap()[val]
		}
		val, found = pairs["ハイスコア時のダンスレベル"]
		if found {
			if val == "E" {
				score.Lamp = 0
			}
		}

		util.SetStructValues(scoreType, reflect.ValueOf(&score), pairs)
	}

	return score, nil
}

// LoadSongDifficulties will lload allll the difficulty levels for
// a provided song ID.
func LoadSongDifficulties(client *http.Client, id *string) ([]int8, error) {
	const musicDetail = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/music_detail.html?index={id}&diff=0"
	result := make([]int8, 0)

	musicDiffDetails := strings.Replace(musicDetail, "{id}", *id, -1)
	res, err := client.Get(musicDiffDetails)

	if err != nil {
		fmt.Print(err)
		return result, err
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

	doc.Find("li.step").Each(func(i int, s *goquery.Selection) {
		img, exists := s.Find("img").Attr("src")
		if exists {
			imgExp := regexp.MustCompile(`songdetails_level_[0-9]*\.png`)
			lvlExp := regexp.MustCompile("[^0-9]+")
			s := imgExp.FindString(img)
			s = lvlExp.ReplaceAllString(s, "")
			v, _ := strconv.Atoi(s)
			result = append(result, int8(v))
		}
	})
	return result, nil
}

func printScores(charts map[string]Chart) {
	for k, v := range charts {
		fmt.Printf("Scores for chart ID %s\n", k)
		vType := reflect.TypeOf(v)
		for i := 0; i < vType.NumField(); i++ {

			score := (reflect.ValueOf(&v).Elem().FieldByIndex([]int{i})).Interface().(Score)
			//fmt.Println(reflect.ValueOf(&v).Type().FieldByIndex([]int{i}).Name)
			fmt.Printf("Level: %d\nScore: %d\nMax Combo: %d\nTotal Plays: %d\nClear Count: %d\nLast Played: %s\nLamp: %d\n\n", score.Level, score.Score, score.MaxCombo, score.PlayCount, score.ClearCount, score.LastPlayed.String(), score.Lamp)
		}
	}
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