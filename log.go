package main

import (
	"bytes"
	"encoding/json"
	"errors"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type DingResponse struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type DingHook struct {
	Webhook *url.URL // 钉钉机器人的 Webhook URL
	client  *http.Client
}

func NewDingHook(webhook string, client *http.Client) (*DingHook, error) {
	wh, err := url.Parse(webhook)
	if err != nil {
		return nil, errors.New("Parse webhook to url.URL error: " + err.Error())
	}

	dh := &DingHook{Webhook: wh}
	if client != nil {
		dh.client = client
	} else {
		dh.client = &http.Client{}
	}

	return dh, err
}

func (dh *DingHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.WarnLevel,
		//logrus.ErrorLevel,
		//logrus.FatalLevel,
		//logrus.PanicLevel,
	}
}

func (dh *DingHook) Fire(entry *logrus.Entry) error {
	b, err := json.Marshal(entry.Data)
	if err != nil {
		return errors.New("Marshal Fields to JSON error: " + err.Error())
	}

	body := ioutil.NopCloser(bytes.NewBuffer(b))
	request := &http.Request{
		Method:     "POST",
		URL:        dh.Webhook,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       body,
		Host:       dh.Webhook.Host,
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := dh.client.Do(request)
	if err != nil {
		return errors.New("Send to DingTalk error: " + err.Error())
	}
	defer response.Body.Close()

	rb, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New("Read DingTalk response error: " + err.Error())
	}

	dr := &DingResponse{}
	err = json.Unmarshal(rb, dr)
	if err != nil {
		return errors.New("Unmarshal DingTalk response to JSON error: " + err.Error())
	}

	if dr.ErrCode != 0 {
		return errors.New("DingTalk return error message: " + dr.ErrMsg)
	}

	return nil
}

func InitLog() {
	// pro
	//dh, _ := NewDingHook("https://oapi.dingtalk.com/robot/send?access_token=8a9cac59c981b6ecfbc6e333ad38652d59f70dc21821e934668b06da5c6f2f98", nil)
	// dev
	//dh, _ := NewDingHook("https://oapi.dingtalk.com/robot/send?access_token=3f4372bd2a832f8fdf070be553bd0b56714a407f5db9f57cceaaf76540adeb5a", nil)
	//logrus.AddHook(dh)

	//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
	//logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	logrus.SetOutput(os.Stdout)
	//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
	file, err := os.OpenFile("log/"+time.Now().Format("20060102")+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	writers := []io.Writer{
		file,
		os.Stdout}
	//同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(writers...)
	if err == nil {
		logrus.SetOutput(fileAndStdoutWriter)
	} else {
		logrus.Error("failed to log to file.")
	}
	//设置最低loglevel
	logrus.SetLevel(logrus.InfoLevel)
}
