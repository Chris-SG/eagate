package user

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chris-sg/eagate/util"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

// Captcha defines Konami Captcha JSON
type Captcha struct {
	Data struct {
		CorrectPic string `json:"correct_pic"`
		Kcsess     string `json:"kcsess"`
		ChoiceList []struct {
			Attr   string `json:"attr"`
			ImgURL string `json:"img_url"`
			Key    string `json:"key"`
		} `json:"choicelist"`
	} `json:"data"`
}

// LoginForm defines Konami Login Form
type LoginForm struct {
	LoginID         string `json:"login_id"`
	Password        string `json:"pass_word"`
	OneTimePassword string `json:"otp,omitempty"`
	Href            string `json:"resrv_url,omitempty"`
	Captcha         string `json:"captcha"`
}

func getChecksums() map[string]string {
	return map[string]string{
		"0753041e08cfa0b322182a1c90647f42": "bomberman",
		"3aea602a1fb82b86df00832599e99550": "bomberman",
		"4f303172009ec741ad86fc08646245a0": "bomberman",
		"71d44daa0af15ee66d35df6b6a55f73f": "bomberman",
		"a0c115ed425765d3ceb36b1bec8c5c5f": "bomberman",
		"b4df277bc8057f92257e06f94c1d6321": "bomberman",
		"be70c62be274742771ad0d6198b918d3": "bomberman",

		"1125bf23bfd6f8ceab1fe91ae1cf2a54": "goemon",
		"222114e728406f87493b88ae388a4440": "goemon",
		"39eb3b335122337be84d3b591855966c": "goemon",
		"93c6c1721f6f8475c502a1237f3c99d5": "goemon",
		"964810d74f7cae311dfe39f2f505b3ef": "goemon",
		"96ee4dbc156c6a8948dde32783d852c8": "goemon",
		"f4d24a8cec9c334318d6590a86e6c349": "goemon",

		"26000240431f6ec78bfcb76dd493a0d9": "twinbee",
		"32908c88748bc6097129c67255a28ae8": "twinbee",
		"5de00f3d6d6d8d5b378da307bc1245f9": "twinbee",
		"9bc7769709fb3ea9dfde557114a4fafc": "twinbee",
		"bfb0932cce45d2bfa132ca05367309e8": "twinbee",

		"272a461ba9b6f69ce6de8ef0670b2679": "shiori",
		"35bc09b719c0ff075097d94733fd712e": "shiori",
		"5cd8a16750d3193d91b97ff26d9805f9": "shiori",
		"9eeddbe6cc1c4e32765ddfb5dafc309c": "shiori",
		"be03cce419a71f81fc43af1c71f47cc1": "shiori",
		"eaa1c324db9a0db9e782ecf7cfffdf4f": "shiori",

		"37116d63462ef81fe3b72ec57ede3f13": "louie",
		"39d0b10cb1fd3f12d310efd07323b873": "louie",
		"568bce3bcb8c5a19fefca1c50934b12e": "louie",
		"7d3132668987d9155363bab78d27e7f4": "louie",
		"b12704c2939f55b66f47d53f6d61af15": "louie",
		"ce734d5ec8b46a7a333609e6dd864f93": "louie",
	}
}

// AddCookiesToJar will add a slice of cookies to the
// provided jar under the eagate URL
func AddCookiesToJar(jar http.CookieJar, cookies []*http.Cookie) {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	jar.SetCookies(eagate, cookies)
}

// GetCookieFromEaGate will submit a request to login as the given
// username with the provided password. This does not yet
// support a OTP.
func GetCookieFromEaGate(username string, password string, client *http.Client) (*http.Cookie, error) {
		const eagateLoginAuthResource = "/gate/p/common/login/api/login_auth.html"

		eagateLoginAuthURI := util.BuildEaURI(eagateLoginAuthResource)

		session, correct, err := SolveCaptcha(client)

		if err != nil {
			return nil, err
		}

		form := url.Values{}

		captchaResult := "k_" + session + correct
		form.Add("login_id", username)
		form.Add("pass_word", password)
		form.Add("captcha", captchaResult)

		//res, err := http.NewRequest("POST", eagateLoginAuth, strings.NewReader(form.Encode()))
		res, err := client.PostForm(eagateLoginAuthURI, form)

		if err != nil {
			return nil, err
		}

		cookies := res.Cookies()

		if len(cookies) == 0 {
			return nil, fmt.Errorf("could not generate cookie")
		}

	return cookies[0], nil
}

func CheckCookieEaGateAccess(client *http.Client, cookie *http.Cookie) error {
	clientJar := client.Jar
	tempJar, _ := cookiejar.New(nil)
	AddCookiesToJar(tempJar, []*http.Cookie{cookie})
	client.Jar = tempJar
	res, err := client.Get("https://p.eagate.573.jp/gate/p/mypage/index.html")
	client.Jar = clientJar
	if err != nil || res.StatusCode != 200 {
		return fmt.Errorf("cookie is no longer valid")
	}
	return nil
}

// SolveCaptcha will load a Konami Captcha and attempt to solve it.
// It returns a string containing the captcha session, a slice containing
// all correct keys, and any errors encountered.
func SolveCaptcha(client *http.Client) (string, string, error) {
	const eagateCaptchaGenerateResource = "/gate/p/common/login/api/kcaptcha_generate.html"

	eagateCaptchaGenerateURI := util.BuildEaURI(eagateCaptchaGenerateResource)

	res, err := client.Get(eagateCaptchaGenerateURI)
	if err != nil {
		return "", "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	var captchaData Captcha

	err = json.Unmarshal([]byte(body), &captchaData)
	if err != nil {
		return "", "", err
	}

	correctPicMD5, err := LoadMD5OfImageURI(captchaData.Data.CorrectPic)
	if err != nil {
		return "", "", err
	}

	correctCharacter, err := FindMD5(correctPicMD5)
	if err != nil {
		re := regexp.MustCompile("[A-Fa-f0-9]{32}")
		match := re.FindStringSubmatch(captchaData.Data.CorrectPic)

		return "", "", fmt.Errorf("character key %s md5 %s was not found", match[0], correctPicMD5)
	}

	type Choice struct {
		md5 string
		key string
	}

	var choiceImages []Choice

	for _, element := range captchaData.Data.ChoiceList {
		picture, err := LoadMD5OfImageURI(element.ImgURL)
		if err == nil {
			choiceImages = append(choiceImages, Choice{picture, element.Key})
		}
	}

	var captchaString string

	for _, element := range choiceImages {
		captchaString += "_"
		character, err := FindMD5(element.md5)
		if err != nil {
			fmt.Println(err)
		} else if character == correctCharacter {
			captchaString += element.key
		} else if character == "unknown" {
			fmt.Printf("%s %s %s", element.key, element.md5, character)
		}
	}

	return captchaData.Data.Kcsess, captchaString, nil
}

// LoadMD5OfImageURI will attempt to Get an image from the provided
// URI, and calculate the MD5 checksum of this image.
// Returns the MD5 checksum as a string and an error if the process fails.
func LoadMD5OfImageURI(uri string) (string, error) {
	image, err := http.Get(uri)

	if err != nil {
		return "", err
	}

	defer image.Body.Close()
	imageData, err := ioutil.ReadAll(image.Body)

	if err != nil {
		return "", err
	}

	correctPicMD5 := md5.Sum([]byte(imageData))

	return fmt.Sprintf("%x", correctPicMD5), nil
}

// FindMD5 will attempt to locate the Captcha MD5 in the existing
// MD5 slices.
// Returns the character name or unknown and an error if not found.
func FindMD5(md5 string) (string, error) {
	if val, ok := getChecksums()[md5]; ok {
		return val, nil
	}

	return "unknown", errors.New("failed to locate captcha type")
}
