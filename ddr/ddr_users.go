package ddr

import (
	"fmt"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_models/ddr_models"
	"reflect"
	"strconv"
	"strings"
	"sync"
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
		fmt.Println(err)
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), sougouDetails)
	util.SetStructValues(pcType, reflect.ValueOf(&pc), sougouDetails)

	fmt.Println(single.Html())

	singleDetails, err := util.TableThTd(single.Find("table.small_table").First())
	if err != nil {
		fmt.Println(err)
	}
	singleMap := make(map[string]string)
	for k, v := range singleDetails {
		singleMap[k + "_single"] = v
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), singleMap)
	util.SetStructValues(pcType, reflect.ValueOf(&pc), singleMap)

	doubleDetails, err := util.TableThTd(double.Find("table.small_table").First())
	if err != nil {
		fmt.Println(err)
	}
	doubleMap := make(map[string]string)
	for k, v := range doubleDetails {
		doubleMap[k + "_double"] = v
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), doubleMap)
	util.SetStructValues(pcType, reflect.ValueOf(&pc), doubleMap)

	fmt.Println(client.Username)
	pi.EaGateUser = &client.Username
	pc.PlayerCode = pi.Code

	fmt.Println(pi)
	fmt.Println(pc)

	return &pi, &pc, nil
}

func SongStatistics(client util.EaClient, charts []ddr_models.SongDifficulty) ([]ddr_models.SongStatistics, error) {
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
			fmt.Println(stat)

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