package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func IsMaintenanceMode(client EaClient) bool {
	doc, err := GetPageContentAsGoQuery(client.Client, "https://p.eagate.573.jp/game/")
	if err != nil {
		return true
	}
	html, _ := doc.Html()
	return strings.Contains(html, "メンテナンス期間")
}

// Find will locate the existence of a given value in a slice.
// Returns the index of the value and whether the value was found.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}

	return -1, false
}

// ShiftJISStringToUTF8String will convert a SHIFT-JIS encoded string into
// a UTF-8 encoded string.
func ShiftJISStringToUTF8String(text string) string {
	var b bytes.Buffer
	cvt := transform.NewWriter(&b, japanese.ShiftJIS.NewDecoder())
	cvt.Write([]byte(text))
	cvt.Close()

	return b.String()
}

// ShiftJISBytesToUTF8Bytes will convert a SHIFT-JIS encoded string into
// a UTF-8 encoded string.
func ShiftJISBytesToUTF8Bytes(text []byte) []byte {
	var b bytes.Buffer
	cvt := transform.NewWriter(&b, japanese.ShiftJIS.NewDecoder())
	cvt.Write(text)
	cvt.Close()

	return b.Bytes()
}

// SetStructValues will set the values of a struct based on a "tag"
// using data provided in the map.
func SetStructValues(structType reflect.Type, structValue reflect.Value, data map[string]string) {
	for i := 0; i < structType.NumField(); i++ {
		tag, exists := structType.Field(i).Tag.Lookup("tag")
		if exists {
			val, found := data[tag]
			if found {
				switch structType.Field(i).Type.String() {
				case "string":
					structValue.Elem().FieldByIndex([]int{i}).SetString(val)
				case "int", "int8":
					reg, _ := regexp.Compile("[^0-9]+")
					val, err := strconv.Atoi(reg.ReplaceAllString(val, ""))
					if err == nil {
						structValue.Elem().FieldByIndex([]int{i}).SetInt(int64(val))
					} else {
						fmt.Printf("Warning: Failed to set int value for %s\n", structType.Field(i).Name)
					}
				case "time.Time":
					format := "2006-01-02 15:04:05"
					loc, err := time.LoadLocation("Asia/Tokyo")
					if err != nil {
						fmt.Println(err)
					}
					t, err := time.ParseInLocation(format, val, loc)
					if err == nil {
						structValue.Elem().FieldByIndex([]int{i}).Set(reflect.ValueOf(t))
					}
				default:
					fmt.Printf("unhandled type %s", structType.Field(i).Type.String())
				}
			}
		}
	}
}

// TableThTd will attempt to separate th/td fields of a table
// into key:value pairs.
func TableThTd(selection *goquery.Selection) (map[string]string, error) {
	if selection.Is("table") {
		results := make(map[string]string)
		tableValues := selection.Find("th, td")
		var th string
		var td string
		for idx := 0; idx < tableValues.Length(); idx++ {
			currSelection := tableValues.Eq(idx)
			if currSelection.Is("th") {
				th = strings.TrimSpace(currSelection.Text())
			} else if currSelection.Is("td") {
				td = strings.TrimSpace(currSelection.Text())
				if th != "" && td != "" {
					results[th] = td
				}
				th = ""
				td = ""
			} else {
				th = ""
				td = ""
			}
		}
		return results, nil
	}
	return make(map[string]string), fmt.Errorf("query selection is not of type table")
}

func GetPageContentAsGoQuery(client *http.Client, resource string) (*goquery.Document, error) {
	res, err := client.Get(resource)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = ShiftJISBytesToUTF8Bytes(body)
		}
	}

	return goquery.NewDocumentFromReader(bytes.NewReader(body))
}

func BuildEaURI(resource string) string {
	const ea = "https://p.eagate.573.jp"
	return ea + resource
}