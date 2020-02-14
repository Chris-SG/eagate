package ddr

import (
	"fmt"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_models/ddr_models"
	"net/http"
	"reflect"
)

func ddrGetLampMap() map[string]int8 {
	return map[string]int8{
		"Failed":      0,
		"---":         1,
		"グッドフルコンボ":    2,
		"グレートフルコンボ":   3,
		"パーフェクトフルコンボ": 4,
	}
}

// PlayerInformation retrieves the base player
// information using the provided cookie.
func PlayerInformation(client *http.Client) (*ddr_models.PlayerDetails, error) {
	const playerInformationResource = "/game/ddr/ddra20/p/playdata/index.html"

	playerInformationURI := util.BuildEaURI(playerInformationResource)
	doc, err := util.GetPageContentAsGoQuery(client, playerInformationURI)
	if err != nil {
		return nil, err
	}

	pi := ddr_models.PlayerDetails{}
	piType := reflect.TypeOf(pi)

	//pi.Name = doc.Find("div#")

	sougou := doc.Find("div#sougou").First()
	single := doc.Find("div#single").First()
	double := doc.Find("div#double").First()

	if sougou == nil || single == nil || double == nil {
		return nil, fmt.Errorf("unable to find all divs")
	}

	sougouDetails, err := util.TableThTd(sougou.Find("table#status").First())
	if err != nil {
		fmt.Println(err)
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), sougouDetails)

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

	doubleDetails, err := util.TableThTd(double.Find("table.small_table").First())
	if err != nil {
		fmt.Println(err)
	}
	doubleMap := make(map[string]string)
	for k, v := range doubleDetails {
		doubleMap[k + "_double"] = v
	}
	util.SetStructValues(piType, reflect.ValueOf(&pi), doubleMap)



	/*doc.Find("div").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")

		if exists {
			_, found := util.Find([]string{"sougou", "single", "double"}, id)
			if found {
				table := s.Find("table").First()
				data, err := util.TableThTd(table)
				if err != nil {
					log.Fatal(err)
				}
				util.SetStructValues(piType, reflect.ValueOf(&pi), data)
			}
		}
	})*/

	fmt.Println(pi)

	return &pi, nil
}