package user

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/chris-sg/eagate/util"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
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

// getChecksums provides a mechanism to map checksums of
// predetermined images to character names for matching
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

// GetCookieFromEaGate will submit a request to login as the given
// username with the provided password and optionally, otp.
func GetCookieFromEaGate(username string, password string, otp string, client util.EaClient) (*http.Cookie, error) {
	glog.Infof("attempting to login user %s", username)
	const eagateLoginAuthResource = "/gate/p/common/login/api/login_auth.html"

	eagateLoginAuthURI := util.BuildEaURI(eagateLoginAuthResource)

	glog.Infof("loading captcha data for user %s", client.GetUsername())
	captchaData, err := LoadCaptchaData(client)
	if err != nil {
		glog.Errorf("user %s failed loading captcha: %s", client.GetUsername(), err.Error())
		return nil, fmt.Errorf("user %s failed to get cookie from eagate", client.GetUsername())
	}

	glog.Infof("solving captcha for user %s", client.GetUsername())
	session, correct, err := SolveCaptcha(captchaData)
	if err != nil {
		glog.Errorf("user %s failed solving captcha: %s", client.GetUsername(), err.Error())
		return nil, fmt.Errorf("user %s failed to get cookie from eagate", client.GetUsername())
	}

	form := url.Values{}

	captchaResult := "k_" + session + correct
	form.Add("login_id", username)
	form.Add("pass_word", password)
	if len(otp) > 0 {
		form.Add("otp", otp)
	}
	form.Add("captcha", captchaResult)

	res, err := client.Client.PostForm(eagateLoginAuthURI, form)

	if err != nil {
		glog.Warningf("user %s failed login: %s", username, err.Error())
		return nil, err
	}

	cookies := res.Cookies()

	if len(cookies) == 0 {
		glog.Errorf("cookie was not generated for user %s", username)
		return nil, fmt.Errorf("could not generate cookie")
	}

	return cookies[0], nil
}

func LoadCaptchaData(client util.EaClient) (captchaData Captcha, err error) {
	const eagateCaptchaGenerateResource = "/gate/p/common/login/api/kcaptcha_generate.html"
	eagateCaptchaGenerateURI := util.BuildEaURI(eagateCaptchaGenerateResource)

	res, err := client.Client.Get(eagateCaptchaGenerateURI)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &captchaData)
	return
}

// SolveCaptcha will load a Konami Captcha and attempt to solve it.
// It returns a string containing the captcha session, a slice containing
// all correct keys, and any errors encountered.
func SolveCaptcha(captchaData Captcha) (session string, correct string, err error) {
	correctPicData, err := LoadImageDataFromUri(captchaData.Data.CorrectPic)
	if err != nil {
		return
	}

	correctPicMD5 := GetMD5FromImageData(correctPicData)

	correctCharacter, err := FindCharacterFromMD5(string(correctPicMD5))
	if err != nil {
		re := regexp.MustCompile("[A-Fa-f0-9]{32}")
		match := re.FindStringSubmatch(captchaData.Data.CorrectPic)

		glog.Errorf("captcha failed due to missing character key %s with md5 %s", match[0], correctPicMD5)
		return "", "", fmt.Errorf("character key %s md5 %s was not found", match[0], correctPicMD5)
	}

	type Choice struct {
		md5 string
		key string
	}

	var choiceImages []Choice

	for _, element := range captchaData.Data.ChoiceList {
		if len(element.ImgURL) == 0 {
			continue
		}
		glog.Infoln(element)
		picture, err := LoadImageDataFromUri(element.ImgURL)
		if err != nil {
			glog.Errorf("could not load image data for url %s: %s\n", element.ImgURL, err.Error())
			continue
		}
		md5 := GetMD5FromImageData(picture)
		if err != nil {
			glog.Errorf("failed to find md5 for url %s: %s", element.ImgURL, err.Error())
			continue
		}
		choiceImages = append(choiceImages, Choice{string(md5), element.Key})
	}

	var captchaString string

	for _, element := range choiceImages {
		captchaString += "_"
		character, err := FindCharacterFromMD5(element.md5)
		if character == correctCharacter {
			captchaString += element.key
		} else if err != nil {
			glog.Errorf("captcha error: %s", err.Error())
		}
	}

	return captchaData.Data.Kcsess, captchaString, nil
}

// LoadMD5OfImageURI will attempt to Get an image from the provided
// URI, and calculate the MD5 checksum of this image.
// Returns the MD5 checksum as a string and an error if the process fails.
func LoadImageDataFromUri(uri string) ([]byte, error) {
	image, err := http.Get(uri)

	if err != nil {
		glog.Errorf("failed to load %s: %s", uri, err.Error())
		return nil, err
	}

	defer image.Body.Close()
	imageData, err := ioutil.ReadAll(image.Body)

	if err != nil {
		glog.Errorf("failed to load %s: %s", uri, err.Error())
		return nil, err
	}
	return imageData, nil
}

// GetMD5FromImageData
func GetMD5FromImageData(imageData []byte) []byte {
	md5 := md5.Sum(imageData)
	return []byte(fmt.Sprintf("%x", md5))
}

// FindCharacterFromMD5 will attempt to locate the Captcha MD5 in the existing
// MD5 slices.
// Returns the character name or unknown and an error if not found.
func FindCharacterFromMD5(md5 string) (string, error) {
	if val, ok := getChecksums()[string(md5)]; ok {
		return val, nil
	}
	return "", fmt.Errorf("failed to locate character for md5 %s", md5)
}