package main

import (
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var (
	apiKey    string
	secretKey string
	oauthUrl  string
	ocrUrl    string

	indexUrl        string
	loginUrl        string
	validateUrl     string
	setParameterUrl string
	downloadUrl     string

	account Account
	path    string
	byType  string
)

type Account struct {
	UserID   string
	Password string
}

func init() {
	InitLog()
	cfg, err := ini.Load("./config/config.ini")
	if err != nil {
		logrus.Error(err)
	}
	apiKey = cfg.Section("ocr").Key("apiKey").String()
	secretKey = cfg.Section("ocr").Key("secretKey").String()
	oauthUrl = cfg.Section("ocr").Key("oauthUrl").String()
	ocrUrl = cfg.Section("ocr").Key("ocrUrl").String()
	if apiKey == "" || secretKey == "" {
		logrus.Warn("请配置API Key和Secret Key")
	}

	indexUrl = cfg.Section("url").Key("indexUrl").String()
	loginUrl = cfg.Section("url").Key("loginUrl").String()
	validateUrl = cfg.Section("url").Key("validateUrl").String() + strconv.Itoa(int(time.Now().UnixNano()/1e6))
	setParameterUrl = cfg.Section("url").Key("setParameterUrl").String()
	downloadUrl = cfg.Section("url").Key("downloadUrl").String()

	account.UserID = cfg.Section("account").Key("userID").String()
	account.Password = cfg.Section("account").Key("password").String()
	if account.UserID == "" || account.Password == "" {
		logrus.Warn("请配置用户名和密码")
	}

	path = cfg.Section("basic").Key("path").String()
	byType = cfg.Section("basic").Key("byType").String()

	if path == "" {
		path = "./"
	}
	if byType == "" {
		byType = "date"
	}
}
