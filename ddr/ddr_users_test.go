package ddr

import (
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/golang/glog"
	"testing"
	"time"
)

func TestPlayerInformationFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/player/index.html"

	// Setup expected results

	expectedPlayerInformation := ddr_models.PlayerDetails{
		Code:        12345678,
		Name:        "EAGATE",
		Prefecture:  "オーストラリア",
		SingleRank:  "段位なし",
		DoubleRank:  "段位なし",
		Affiliation: "所属なし",
	}

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	playerInformation, err := playerInformationFromPlayerDocument(document)
	if err != nil {
		t.Fatalf("error in playerInformationFromPlayerDocument: %s", err.Error())
	}

	if  playerInformation.Code != expectedPlayerInformation.Code ||
		playerInformation.Name != expectedPlayerInformation.Name ||
		playerInformation.Prefecture != expectedPlayerInformation.Prefecture ||
		playerInformation.SingleRank != expectedPlayerInformation.SingleRank ||
		playerInformation.DoubleRank != expectedPlayerInformation.DoubleRank ||
		playerInformation.Affiliation != expectedPlayerInformation.Affiliation {
		t.Errorf("player information did not match, expected %+#v but got %+#v", expectedPlayerInformation, playerInformation)
	}
}

func TestPlaycountFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/player/index.html"

	// Setup expected results

	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	expectedLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedSingleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedDoubleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-02-20 19:11:10", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}

	expectedPlaycount := ddr_models.Playcount{
		Playcount:          380,
		LastPlayDate:       expectedLastPlayDate,
		SinglePlaycount:    360,
		SingleLastPlayDate: expectedSingleLastPlayDate,
		DoublePlaycount:    20,
		DoubleLastPlayDate: expectedDoubleLastPlayDate,
		Player:             ddr_models.PlayerDetails{},
		PlayerCode:         12345678,
	}

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	playcount, err := playcountFromPlayerDocument(document)
	if err != nil {
		t.Fatalf("error in playerInformationFromPlayerDocument: %s", err.Error())
	}

	if playcount.Playcount != expectedPlaycount.Playcount ||
		playcount.LastPlayDate.String() != expectedPlaycount.LastPlayDate.String() ||
		playcount.SinglePlaycount != expectedPlaycount.SinglePlaycount ||
		playcount.SingleLastPlayDate.String() != expectedPlaycount.SingleLastPlayDate.String() ||
		playcount.DoublePlaycount != expectedPlaycount.DoublePlaycount ||
		playcount.DoubleLastPlayDate.String() != expectedPlaycount.DoubleLastPlayDate.String() ||
		playcount.PlayerCode != expectedPlaycount.PlayerCode {
		t.Errorf("playcount information did not match, expected %+#v but got %+#v", expectedPlaycount, playcount)
	}
}

func TestPlayerInformationForClient(t *testing.T) {
	// Setup test
	const playerProfilePage = "./test_data/player/index.html"
	const playerProfileUri = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/index.html"
	uriMapping := make(map[string]string)
	uriMapping[playerProfileUri] = playerProfilePage

	c, s := testServerAndClient(uriMapping)
	defer s.Close()

	// Setup expected results

	expectedPlayerInformation := ddr_models.PlayerDetails{
		Code:        12345678,
		Name:        "EAGATE",
		Prefecture:  "オーストラリア",
		SingleRank:  "段位なし",
		DoubleRank:  "段位なし",
		Affiliation: "所属なし",
	}

	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	expectedLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedSingleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedDoubleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-02-20 19:11:10", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}

	expectedPlaycount := ddr_models.Playcount{
		Playcount:          380,
		LastPlayDate:       expectedLastPlayDate,
		SinglePlaycount:    360,
		SingleLastPlayDate: expectedSingleLastPlayDate,
		DoublePlaycount:    20,
		DoubleLastPlayDate: expectedDoubleLastPlayDate,
		Player:             ddr_models.PlayerDetails{},
		PlayerCode:         12345678,
	}

	// Run test
	playerInformation, playcount, err := PlayerInformationForClient(c)

	if  playerInformation.Code != expectedPlayerInformation.Code ||
		playerInformation.Name != expectedPlayerInformation.Name ||
		playerInformation.Prefecture != expectedPlayerInformation.Prefecture ||
		playerInformation.SingleRank != expectedPlayerInformation.SingleRank ||
		playerInformation.DoubleRank != expectedPlayerInformation.DoubleRank ||
		playerInformation.Affiliation != expectedPlayerInformation.Affiliation {
		t.Errorf("player information did not match, expected %+#v but got %+#v", expectedPlayerInformation, playerInformation)
	}

	if playcount.Playcount != expectedPlaycount.Playcount ||
		playcount.LastPlayDate.String() != expectedPlaycount.LastPlayDate.String() ||
		playcount.SinglePlaycount != expectedPlaycount.SinglePlaycount ||
		playcount.SingleLastPlayDate.String() != expectedPlaycount.SingleLastPlayDate.String() ||
		playcount.DoublePlaycount != expectedPlaycount.DoublePlaycount ||
		playcount.DoubleLastPlayDate.String() != expectedPlaycount.DoubleLastPlayDate.String() ||
		playcount.PlayerCode != expectedPlaycount.PlayerCode {
		t.Errorf("playcount information did not match, expected %+#v but got %+#v", expectedPlaycount, playcount)
	}
}

func TestSongStatisticsFromDocument(t *testing.T) {

}

func TestSongStatisticsForClient(t *testing.T) {

}

func TestRecentScoresFromDocument(t *testing.T) {

}

func TestRecentScoresForClient(t *testing.T) {

}

func TestWorkoutDataFromDocument(t *testing.T) {

}

func TestWorkoutDataForClient(t *testing.T) {

}