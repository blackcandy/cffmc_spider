package main

// @Title        ocr.go
// @Description  图形验证码识别
// @Create       Spenser 2024-05-23

import (
	"encoding/base64"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type tokenJson struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	Scope         string `json:"scope"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	SessionSecret string `json:"session_secret"`
}

// 返回参数
type wordResult struct {
	LogId          int `json:"log_id"`
	WordsResultNum int `json:"words_result_num"`
	WordsResult    []struct {
		Words string `json:"words"`
	} `json:"words_result"`
}

//var apiKey string = "KS0jcHyqruAWEOeh9wpZlgah"
//var secretKey string = "TSZ7dZswgk0Xai6O2XMNdGUVekS7yy1i"
//var oauthUrl string = "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials"
//var ocrUrl string = "https://aip.baidubce.com/rest/2.0/ocr/v1/accurate_basic"

// getAccessToken
//
//	@Description: 获取access_token
//	@return accessToken
//	@return err
func getAccessToken() (accessToken string, err error) { //
	host := oauthUrl + "&client_id=" + apiKey + "&client_secret=" + secretKey + "&"
	resp, err := http.Get(host)
	if err != nil {
		return "", err
	}
	var t tokenJson
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

// GetCodeByBase64
//
// @Description: 根据图像base64识别二维码
// @param imgUrl
// @param cookie
// @return code
// @return err
func GetCodeByBase64(imgUrl string, cookie string) (code string) {
	accessToken, err := getAccessToken()
	if err != nil {
		return ""
	}
	uri, err := url.Parse(ocrUrl)
	if err != nil {
		return ""
	}
	query := uri.Query()
	query.Set("access_token", accessToken)
	uri.RawQuery = query.Encode()

	client := &http.Client{}
	req, err := http.NewRequest("GET", imgUrl, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept-Encoding", "gzip, deflate, sdch")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36")
	req.Header.Set("Accept", "text/javascript, text/html, application/xml, text/xml, */*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	image := base64.StdEncoding.EncodeToString(fileBytes)

	sendBody := http.Request{}
	sendBody.ParseForm()
	sendBody.Form.Add("image", image)
	sendBody.Form.Add("language_type", "CHN_ENG")
	sendData := sendBody.Form.Encode()

	client2 := &http.Client{}
	request, err := http.NewRequest("POST", uri.String(), strings.NewReader(sendData))
	if err != nil {
		return ""
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//request.Header.Set("Cookie", cookie)

	response, err := client2.Do(request)
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ""
	}
	//fmt.Println(string(result))
	var w wordResult
	err = json.Unmarshal([]byte(string(result)), &w)
	if err != nil {
		logrus.Error(err)
	}

	return w.WordsResult[0].Words
}
