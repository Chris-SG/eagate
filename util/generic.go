package util

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// GenerateClient will generate a http.client that is
// used by this library.
func GenerateClient() *http.Client {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	return client
}

// CheckJarForCookies will search a client's cookiejar for cookies.
func CheckJarForCookies(client *http.Client) bool {
	pigate, _ := url.Parse("https://p.eagate.573.jp/")

	c2 := client.Jar.Cookies(pigate)

	for _, element := range c2 {
		fmt.Printf("%s: %s from c2", element.Name, element.Value)
	}
	return true
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
					t, err := time.Parse(format, val)
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
				th = currSelection.Text()
			} else if currSelection.Is("td") {
				td = currSelection.Text()
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
