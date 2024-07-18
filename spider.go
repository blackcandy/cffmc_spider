package main

import (
	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

var cookie, validateCode, token string

func DownloadSettlementDocument(account Account, tradeDate string, byType string, path string) (bool, error) {
	done := make(chan bool)
	var err error

	go func(done chan bool) {
		c := colly.NewCollector(
			colly.MaxBodySize(100*1024*1024),
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36"),
			colly.MaxDepth(1),
			colly.Async(true),
			//colly.DetectCharset(),
			colly.AllowURLRevisit(),
		)

		c.WithTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   90 * time.Second,
				KeepAlive: 90 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   90 * time.Second,
			ExpectContinueTimeout: 20 * time.Second,
		})
		c.SetRequestTimeout(90 * time.Second)

		// 请求头
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Host", indexUrl)
			r.Headers.Set("Connection", "keep-alive")
			r.Headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
			r.Headers.Set("Cookie", cookie)
			r.Headers.Set("Origin", indexUrl)
		})

		//	获取error-msg和token
		c.OnHTML("html", func(e *colly.HTMLElement) {
			if errMsg := e.ChildText("span[class|=error-msg]"); errMsg != "" {
				logrus.Warn("登录失败，错误信息：", errMsg)
				done <- false
			}

			if token = e.ChildAttr("input[name|='org.apache.struts.taglib.html.TOKEN']", "value"); token != "" {
				logrus.Info("token: ", token)
			}
		})

		// 响应
		c.OnResponse(func(r *colly.Response) {
			if r.StatusCode != 200 {
				logrus.Warn("response status code is not 200")
				done <- false
			}

			// 首页响应
			if r.Request.URL.String() == indexUrl {
				logrus.Info("index response success, set cookie")
				cookie = r.Headers.Get("Set-Cookie")

				// 识别验证码
				for i := 0; i < 5; i++ {
					validateCode = GetCodeByBase64(validateUrl, cookie)
					logrus.Info("验证码：", validateCode)
					if len(validateCode) == 6 {
						break
					}
				}
				//validateCode = ocr.GetCodeByBase64(cmd.ValidateUrl, cookie)
				//logrus.Info("验证码：", validateCode)
				if len(validateCode) != 6 {
					logrus.Warn("验证码识别失败")
					done <- false
				} else {
					// 登录
					err := c.Post(loginUrl, map[string]string{
						"userID":   account.UserID,
						"password": account.Password,
						"vericode": validateCode,
					})
					if err != nil {
						logrus.Warn("登录报错：", err)
						done <- false
					}
				}
			}

			// 登录响应
			if r.Request.URL.String() == loginUrl {
				if r.Headers.Get("Set-Cookie") != "" {
					logrus.Info("login success, set cookie")
					cookie = r.Headers.Get("Set-Cookie")

					// 设置日期
					err := c.Post(setParameterUrl, map[string]string{
						"org.apache.struts.taglib.html.TOKEN": token,
						"tradeDate":                           tradeDate,
						"byType":                              byType, //date:逐日盯市，trade:逐笔对冲
					})
					if err != nil {
						logrus.Warn("设置日期报错：", err)
						done <- false
					}
				} else {
					logrus.Warn("login failed, retry")
					done <- false
				}
			}

			// 设置日期响应
			if r.Request.URL.String() == setParameterUrl {
				logrus.Info("set date " + tradeDate + ", start download")
				// 下载
				err := c.Visit(downloadUrl)
				if err != nil {
					logrus.Warn("下载报错：", err)
					done <- false
				}
			}

			// 下载响应
			if r.Request.URL.String() == downloadUrl {
				logrus.Info("download response success, save file")
				fileName := path + "\\" + account.UserID + "_" + tradeDate + ".xls"
				err := r.Save(fileName)
				if err != nil {
					logrus.Warn("saving file "+fileName+" failed:", err)
					done <- false
				} else {
					logrus.Info("saving file " + fileName + " success")
					done <- true
				}
			}
		})

		// 访问首页
		if err := c.Visit(indexUrl); err != nil {
			logrus.Warn("访问首页报错：", err)
			done <- false
		}

		//err := c.Visit("file://../data/投资者查询服务系统 -中国期货市场监控中心.html")
		//if err != nil {
		//	logrus.Warn("访问报错：", err)
		//}
	}(done)

	res := <-done
	return res, err
}
