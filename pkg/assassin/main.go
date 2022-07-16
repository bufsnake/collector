package assassin

import (
	"crypto/tls"
	"fmt"
	"github.com/bufsnake/collector/config"
	"github.com/bufsnake/collector/pkg/log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// 刺客
type Assassin struct {
	conf       *config.Config
	log        *log.Log
	percentage *log.Percentage // 适用于分片下载
	parts      *sync.Map       // 存储成功的part
}

func NewAssassin(conf *config.Config, log *log.Log) *Assassin {
	return &Assassin{conf: conf, log: log, parts: &sync.Map{}}
}

// 判断是否支持断点续传
// 返回值: 是否，响应体大小，错误
func (a *Assassin) DetectBreakpointContinuingly(urlstr string) (bool, int, error) {
	cli := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(a.conf.Timeout) * time.Second,
	}
	cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequest(http.MethodHead, urlstr, nil)
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("Connection", "close")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	req.Header.Set("X-Originating-Ip", "127.0.0.1")
	req.Header.Set("X-Remote-Ip", "127.0.0.1")
	req.Header.Set("X-Remote-Addr", "127.0.0.1")
	req.Header.Set("cf-connecting-ip", "127.0.0.1")
	res, err := cli.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return false, 0, fmt.Errorf("%s status code is %s", urlstr, res.Status)
	}
	cl := int(res.ContentLength)
	if res.ContentLength == -1 {
		cl, err = strconv.Atoi(res.Header.Get("Content-Length"))
		if err != nil {
			return false, 0, fmt.Errorf("%s get content-length error, content-length: %s", urlstr, res.Header.Get("Content-Length"))
		}
	}
	if res.Header.Get("Accept-Ranges") != "" {
		return true, cl, nil
	}
	return false, cl, nil
}
