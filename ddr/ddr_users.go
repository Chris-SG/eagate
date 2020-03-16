package ddr

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_models/ddr_models"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PlayerInformation retrieves the base player
// information using the provided cookie.
func PlayerInformation(client util.EaClient) (*ddr_models.PlayerDetails, *ddr_models.Playcount, error) {
	const playerInformationResource = "/game/ddr/ddra20/p/playdata/index.html"

	playerInformationURI := util.BuildEaURI(playerInformationResource)
	doc, err := util.GetPageContentAsGoQuery(client.Client, playerInformationURI)
	if err != nil {
		return nil, nil, err
	}

	pi := ddr_models.PlayerDetails{}
	piType := reflect.TypeOf(pi)

	pc := ddr_models.Playcount{}
	pcType := reflect.TypeOf(pc)

	//pi.Name = doc.Find("div#")

	sougou := doc.Find("div#sougou").First()
	single := doc.Find("div#single").First()
	double := doc.Find("div#double").First()

	if sougou == nil || single == nil || double == nil {
		return nil, nil, fmt.Errorf("unable to find all divs")
	}

	sougouDetails, err := util.TableThTd(sougou.Find("table#status").First())
	if err != nil {
		return nil, nil, err
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), sougouDetails)
	util.SetStructValues(pcType, reflect.ValueOf(&pc), sougouDetails)

	singleDetails, err := util.TableThTd(single.Find("table.small_table").First())
	if err != nil {
		return nil, nil, err
	}
	singleMap := make(map[string]string)
	for k, v := range singleDetails {
		singleMap[k + "_single"] = v
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), singleMap)
	util.SetStructValues(pcType, reflect.ValueOf(&pc), singleMap)

	doubleDetails, err := util.TableThTd(double.Find("table.small_table").First())
	if err != nil {
		return nil, nil, err
	}
	doubleMap := make(map[string]string)
	for k, v := range doubleDetails {
		doubleMap[k + "_double"] = v
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), doubleMap)
	util.SetStructValues(pcType, reflect.ValueOf(&pc), doubleMap)

	eagateUser := client.GetUsername()
	pi.EaGateUser = &eagateUser
	pc.PlayerCode = pi.Code

	return &pi, &pc, nil
}

func SongStatistics(client util.EaClient, charts []ddr_models.SongDifficulty, playerCode int) ([]ddr_models.SongStatistics, error) {
	var (
		songMtx = &sync.Mutex{}
		songStatistics = make([]ddr_models.SongStatistics, 0)
	)

	const musicDetailResource = "/game/ddr/ddra20/p/playdata/music_detail.html?index={id}&diff={diff}"

	musicDetailUri := util.BuildEaURI(musicDetailResource)

	wg := new(sync.WaitGroup)
	wg.Add(len(charts))

	errorCount := 0

	for _, chart := range charts {
		go func(diff ddr_models.SongDifficulty, songStats *[]ddr_models.SongStatistics) {
			defer wg.Done()
			chartDetails := strings.Replace(musicDetailUri, "{id}", diff.SongId, -1)
			diffId := int(ddr_models.StringToDifficulty(diff.Difficulty)) + (5 * int(ddr_models.StringToMode(diff.Mode)))
			chartDetails = strings.Replace(chartDetails, "{diff}", strconv.Itoa(diffId), -1)

			doc, err := util.GetPageContentAsGoQuery(client.Client, chartDetails)
			if err != nil {
				errorCount++
				return
			}

			if strings.Contains(doc.Find("div#popup_cnt").Text(), "NO PLAY") {
				return
			}

			details, err := util.TableThTd(doc.Find("table#music_detail_table"))
			if err != nil {
				errorCount++
				return
			}

			stat := ddr_models.SongStatistics{}
			statType := reflect.TypeOf(stat)

			util.SetStructValues(statType, reflect.ValueOf(&stat), details)
			stat.SongId = diff.SongId
			stat.Difficulty = diff.Difficulty
			stat.Mode = diff.Mode
			stat.PlayerCode = playerCode

			songMtx.Lock()
			defer songMtx.Unlock()
			*songStats = append(*songStats, stat)
		}(chart, &songStatistics)
	}

	wg.Wait()

	if errorCount > 0 {
		return songStatistics, fmt.Errorf("Failed getting score data for ")
	}

	return songStatistics, nil
}


func RecentScores(client util.EaClient, playerCode int) (*[]ddr_models.Score, error) {
	const recentSongsResource = "/game/ddr/ddra20/p/playdata/music_recent.html"
	
	recentSongsUri := util.BuildEaURI(recentSongsResource)
	
	recentScores := make([]ddr_models.Score, 0)

	doc, err := util.GetPageContentAsGoQuery(client.Client, recentSongsUri)
	if err != nil {
		return nil, err
	}

	table := doc.Find("table#data_tbl")
	if table.Length() == 0 {
		return nil, fmt.Errorf("could not find data_tbl")
	}


	tableBody := table.First().Find("tbody").First()
	if tableBody == nil {
		return nil, fmt.Errorf("could not find table body")
	}

	tableBody.Find("tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() > 0 {
			var score ddr_models.Score
			info := s.Find("a.music_info.cboxelement").First()
			href, exists := info.Attr("href")
			if exists {
				difficulty, err := strconv.Atoi(href[len(href)-1:])
				if err == nil {
					score.Mode = ddr_models.Mode(difficulty / 5).String()
					score.Difficulty = ddr_models.Difficulty(difficulty % 5).String()
					score.SongId = href[strings.Index(href, "=")+1 : strings.Index(href, "&")]
				}
			}
			score.Score, _ = strconv.Atoi(s.Find("td.score").First().Text())

			format := "2006-01-02 15:04:05"
			timeSelection := s.Find("td.date").First()

			loc, err := time.LoadLocation("Asia/Tokyo")
			t, err := time.ParseInLocation(format, timeSelection.Text(), loc)
			if err == nil {
				score.TimePlayed = t
			}

			rankSelection := s.Find("td.rank").First()
			imgSelection := rankSelection.Find("img").First()
			path, exists := imgSelection.Attr("src")
			if exists {
				score.ClearStatus = !strings.Contains(path, "rank_s_e")
			}

			score.PlayerCode = playerCode

			recentScores = append(recentScores, score)
		}
	})

	return &recentScores, nil
}

func WorkoutData(client util.EaClient, playerCode int) ([]ddr_models.WorkoutData, error) {
	const workoutResource = "/game/ddr/ddra20/p/playdata/workout.html"

	workoutUri := util.BuildEaURI(workoutResource)

	workoutData := make([]ddr_models.WorkoutData, 0)

	doc, err := util.GetPageContentAsGoQuery(client.Client, workoutUri)
	if err != nil {
		return workoutData, err
	}

	table := doc.Find("table#work_out_left")
	if table.Length() == 0 {
		return workoutData, fmt.Errorf("could not find work_out_left")
	}

	tableBody := table.First().Find("tbody").First()
	if tableBody == nil {
		return workoutData, fmt.Errorf("could not find table body")
	}

	tableBody.Find("tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() == 5 {
			wd := ddr_models.WorkoutData{}
			s.Find("td").Each(func(i int, dataSelection *goquery.Selection) {
				if i == 1 {
					format := "2006-01-02"

					loc, err := time.LoadLocation("Asia/Tokyo")
					t, err := time.ParseInLocation(format, dataSelection.Text(), loc)
					if err == nil {
						wd.Date = t
					}
				} else if i == 2 {
					numerical, err := regexp.Compile("[^0-9]+")
					if err != nil {
						panic(err)
					}
					numericStr := numerical.ReplaceAllString(dataSelection.Text(), "")
					wd.PlayCount, _ = strconv.Atoi(numericStr)
				} else if i == 3 {
					numerical, err := regexp.Compile("[^0-9.]+")
					if err != nil {
						panic(err)
					}
					numericStr := numerical.ReplaceAllString(dataSelection.Text(), "")
					kcalFloat, err := strconv.ParseFloat(numericStr, 32)
					if err == nil {
						wd.Kcal = float32(kcalFloat)
					}
				}
			})
			wd.PlayerCode = playerCode
			workoutData = append(workoutData, wd)
		}
	})

	return workoutData, nil
}