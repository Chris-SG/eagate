package ddr

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/golang/glog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func PlayerInformationForClient(client util.EaClient) (playerDetails ddr_models.PlayerDetails, playcount ddr_models.Playcount, err error) {
	document, err := playerInformationDocument(client)
	if err != nil {
		return
	}
	playerDetails, err = playerInformationFromPlayerDocument(document)
	if err != nil {
		return
	}
	playcount, err = playcountFromPlayerDocument(document)
	if err != nil {
		return
	}

	return
}

func playerInformationFromPlayerDocument(document *goquery.Document) (playerDetails ddr_models.PlayerDetails, err error) {
	status := document.Find("table#status").First()
	if status == nil {
		err = fmt.Errorf("cannot find status table")
		return
	}
	statusDetails, err := util.TableThTd(status)
	if err != nil {
		return
	}
	playerDetails.Name = statusDetails["ダンサーネーム"]
	code, err := strconv.ParseInt(statusDetails["DDR-CODE"], 10, 32)
	if err != nil {
		return
	}
	playerDetails.Code = int(code)
	playerDetails.Prefecture = statusDetails["所属都道府県"]
	playerDetails.SingleRank = statusDetails["段位(SINGLE)"]
	playerDetails.DoubleRank = statusDetails["段位(DOUBLE)"]
	playerDetails.Affiliation = statusDetails["所属クラス"]

	return
}

func playcountFromPlayerDocument(document *goquery.Document) (playcount ddr_models.Playcount, err error) {
	status := document.Find("table#status").First()
	if status == nil {
		err = fmt.Errorf("cannot find status table")
		return
	}
	single := document.Find("div#single table.small_table").First()
	if single == nil {
		err = fmt.Errorf("cannot find single table")
		return
	}
	double := document.Find("div#double table.small_table").First()
	if double == nil {
		err = fmt.Errorf("cannot find double table")
		return
	}

	numericalStripper, _ := regexp.Compile("[^0-9]+")
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	statusDetails, err := util.TableThTd(status)
	if err != nil {
		return
	}
	singleDetails, err := util.TableThTd(single)
	if err != nil {
		return
	}
	doubleDetails, err := util.TableThTd(double)
	if err != nil {
		return
	}

	code, err := strconv.ParseInt(statusDetails["DDR-CODE"], 10, 32)
	if err != nil {
		return
	}
	playcount.PlayerCode = int(code)

	playcount.Playcount, err = strconv.Atoi(numericalStripper.ReplaceAllString(statusDetails["総プレー回数"], ""))
	if err != nil {
		return
	}
	playcount.LastPlayDate, err = time.ParseInLocation(timeFormat, statusDetails["最終プレー日時"], timeLocation)
	if err != nil {
		return
	}

	playcount.SinglePlaycount, err = strconv.Atoi(numericalStripper.ReplaceAllString(singleDetails["プレー回数"], ""))
	if err != nil {
		return
	}
	playcount.SingleLastPlayDate, err = time.ParseInLocation(timeFormat, singleDetails["最終プレー日時"], timeLocation)
	if err != nil {
		return
	}

	playcount.DoublePlaycount, err = strconv.Atoi(numericalStripper.ReplaceAllString(doubleDetails["プレー回数"], ""))
	if err != nil {
		return
	}
	playcount.DoubleLastPlayDate, err = time.ParseInLocation(timeFormat, doubleDetails["最終プレー日時"], timeLocation)
	if err != nil {
		return
	}

	return
}

func SongStatistics(client util.EaClient, charts []ddr_models.SongDifficulty, playerCode int) ([]ddr_models.SongStatistics, error) {
	glog.Infof("loading songstatistics for user %s (%d charts)\n", client.GetUsername(), len(charts))
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
			if ddr_models.StringToMode(diff.Mode) == ddr_models.Double {
				diffId--
			}
			chartDetails = strings.Replace(chartDetails, "{diff}", strconv.Itoa(diffId), -1)

			doc, err := util.GetPageContentAsGoQuery(client.Client, chartDetails)
			if err != nil {
				glog.Errorf("failed loading song statistic for %s: %s\n", client.GetUsername(), err.Error())
				errorCount++
				return
			}

			if strings.Contains(doc.Find("div#popup_cnt").Text(), "NO PLAY") {
				return
			}

			details, err := util.TableThTd(doc.Find("table#music_detail_table"))
			if err != nil {
				glog.Errorf("failed loading song statistic for %s: %s\n", client.GetUsername(), err.Error())
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
		glog.Errorf("failed loading song statistic for %s due to %d errors\n", client.GetUsername(), errorCount)
		return songStatistics, fmt.Errorf("Failed getting score data for ")
	}

	glog.Infof("got %d statistics for user %s\n", len(songStatistics), client.GetUsername())
	return songStatistics, nil
}


func RecentScores(client util.EaClient, playerCode int) (*[]ddr_models.Score, error) {
	glog.Infof("loading recentscores for user %s (playercode %d)\n", client.GetUsername(), playerCode)
	const recentSongsResource = "/game/ddr/ddra20/p/playdata/music_recent.html"
	
	recentSongsUri := util.BuildEaURI(recentSongsResource)
	
	recentScores := make([]ddr_models.Score, 0)

	doc, err := util.GetPageContentAsGoQuery(client.Client, recentSongsUri)
	if err != nil {
		glog.Errorf("failed recentscores for %s: %s\n", client.GetUsername(), err.Error())
		return nil, err
	}

	table := doc.Find("table#data_tbl")
	if table.Length() == 0 {
		glog.Errorf("failed recentscores for %s: could not find data_tbl\n", client.GetUsername())
		return nil, fmt.Errorf("could not find data_tbl")
	}


	tableBody := table.First().Find("tbody").First()
	if tableBody == nil {
		glog.Errorf("failed recentscores for %s: could not find table body\n", client.GetUsername())
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
					if ddr_models.StringToMode(score.Mode) == ddr_models.Double {
						difficulty++
					}
					score.Difficulty = ddr_models.Difficulty(difficulty % 5).String()
					score.SongId = href[strings.Index(href, "=")+1 : strings.Index(href, "&")]
				} else {
					glog.Errorf("strconv failed: %s\n", err.Error())
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

	glog.Infof("recentscores loaded for for %s (%d scores)\n", client.GetUsername(), len(recentScores))
	return &recentScores, nil
}

func WorkoutData(client util.EaClient, playerCode int) ([]ddr_models.WorkoutData, error) {
	glog.Infof("loading workoutdata for user %s (playercode %d)\n", client.GetUsername(), playerCode)
	const workoutResource = "/game/ddr/ddra20/p/playdata/workout.html"

	workoutUri := util.BuildEaURI(workoutResource)

	workoutData := make([]ddr_models.WorkoutData, 0)

	doc, err := util.GetPageContentAsGoQuery(client.Client, workoutUri)
	if err != nil {
		glog.Errorf("failed workoutdata for %s: %s\n", client.GetUsername(), err.Error())
		return workoutData, err
	}

	table := doc.Find("table#work_out_left")
	if table.Length() == 0 {
		glog.Errorf("failed workoutdata for %s: could not find work_out_left\n", client.GetUsername())
		return workoutData, fmt.Errorf("could not find work_out_left")
	}

	tableBody := table.First().Find("tbody").First()
	if tableBody == nil {
		glog.Errorf("failed workoutdata for %s: could not find table body\n", client.GetUsername())
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
						glog.Errorf("regex failure! %s\n", err.Error())
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

	glog.Infof("workoutdata loaded for user %s (%d datapoints)\n", client.GetUsername(), len(workoutData))
	return workoutData, nil
}