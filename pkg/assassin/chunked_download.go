package assassin

import (
	"crypto/tls"
	"github.com/bufsnake/collector/pkg/log"
	"github.com/bufsnake/collector/pkg/useragent"
	"github.com/bufsnake/collector/pkg/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// chunked传输
func (a *Assassin) ChunkedDownload(urlstr string) error {
	cli := &http.Client{
		Timeout: 2 * time.Hour,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	tr := http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		// 设置TLS链接和读取响应头的超时
		TLSHandshakeTimeout:   time.Duration(a.conf.Timeout) * time.Second,
		ResponseHeaderTimeout: time.Duration(a.conf.Timeout) * time.Second,
	}
	cli.Transport = &tr
	req, err := http.NewRequest(http.MethodGet, urlstr, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Referer", "http://www.baidu.com")
	req.Header.Add("Connection", "close")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("User-Agent", useragent.RandomUserAgent())
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	req.Header.Set("X-Originating-Ip", "127.0.0.1")
	req.Header.Set("X-Remote-Ip", "127.0.0.1")
	req.Header.Set("X-Remote-Addr", "127.0.0.1")
	req.Header.Set("cf-connecting-ip", "127.0.0.1")
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return a.chunked_save_file(res.Body, urlstr)
}

func (a *Assassin) chunked_save_file(body io.ReadCloser, urlstr string) error {
	outfile, err := os.Create(strings.TrimRight(a.conf.Folder, "/") + "/" + utils.URL2Filename(urlstr))
	if err != nil {
		return err
	}
	defer outfile.Close()

	percentage := log.NewPercentage(utils.URL2Filename(urlstr), 0, "c")
	a.log.Add(percentage)
	downloader := &Downloader{ReadCloser: body, percentage: percentage, chunked: true}
	_, err = io.Copy(outfile, downloader)
	if err != nil {
		return err
	}
	percentage.AddProgress(downloader.count)
	return err
}
