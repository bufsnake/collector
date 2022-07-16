package assassin

import (
	"crypto/tls"
	"fmt"
	"github.com/bufsnake/collector/pkg/log"
	"github.com/bufsnake/collector/pkg/useragent"
	"github.com/bufsnake/collector/pkg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type part_info struct {
	part  int
	start int
	end   int
}

// 支持断点续传文件下载
func (a *Assassin) BreakpointContinuinglyDownload(urlstr string, content_length int) error {
	a.percentage = log.NewPercentage(utils.URL2Filename(urlstr), content_length, "m")
	a.log.Add(a.percentage)
	part_size := 2 * 1024 * 1024
	task_parts := make(map[int][2]int)
	cycle := content_length / part_size
	if content_length%part_size > 0 {
		cycle++
	}
	for i := 0; i < cycle; i++ {
		end := (i + 1) * part_size
		if (i+1)*part_size >= content_length {
			end = content_length + 1
		}
		task_parts[i] = [2]int{i * part_size, end - 1}
	}
	dw := sync.WaitGroup{}
	part_infos := make(chan part_info)
	for i := 0; i < a.conf.Thread; i++ {
		dw.Add(1)
		go a.download(&dw, urlstr, part_infos)
	}
	for i := 0; i < len(task_parts); i++ {
		part_infos <- part_info{part: i, start: task_parts[i][0], end: task_parts[i][1]}
	}
	close(part_infos)
	dw.Wait()
	// 判断是否存在未完成part
	m := 0
	for ; m < a.conf.MaxRetry; m++ {
		time.Sleep(2 * time.Second)
		part_infos = make(chan part_info)
		for i := 0; i < a.conf.Thread; i++ {
			dw.Add(1)
			go a.download(&dw, urlstr, part_infos)
		}
		download := false
		for i := 0; i < len(task_parts); i++ {
			_, ok := a.parts.Load(i)
			if ok {
				continue
			}
			download = true
			a.log.Println(urlstr, fmt.Sprintf("%drd retry", m+1), "part", i)
			// 尝试下载未完成part
			part_infos <- part_info{part: i, start: task_parts[i][0], end: task_parts[i][1]}
		}
		close(part_infos)
		dw.Wait()
		if !download {
			break
		}
	}
	if m == a.conf.MaxRetry {
		return fmt.Errorf("%s download error", urlstr)
	}
	return a.merge(urlstr, len(task_parts))
}

func (a *Assassin) download(wait *sync.WaitGroup, urlstr string, part_infos chan part_info) {
	defer wait.Done()
	for pi := range part_infos {
		cli := &http.Client{
			Timeout: 10 * time.Minute,
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
			a.log.Println("part", pi.part, urlstr, err)
			continue
		}
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Referer", "http://www.baidu.com")
		req.Header.Set("Connection", "close")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("User-Agent", useragent.RandomUserAgent())
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		req.Header.Set("X-Originating-Ip", "127.0.0.1")
		req.Header.Set("X-Remote-Ip", "127.0.0.1")
		req.Header.Set("X-Remote-Addr", "127.0.0.1")
		req.Header.Set("Cf-Connecting-Ip", "127.0.0.1")
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", pi.start, pi.end))
		res, err := cli.Do(req)
		if err != nil {
			a.log.Println("do req part", pi.part, urlstr, err)
			continue
		}
		err = a.savepartfile(res.Body, urlstr+fmt.Sprintf(".%d.part", pi.part))
		res.Body.Close()
		if err != nil {
			a.log.Println("save part file", pi.part, urlstr, err)
			continue
		}
		a.parts.Store(pi.part, true)
	}
}

func (a *Assassin) savepartfile(body io.ReadCloser, urlstr string) error {
	part_file := "/tmp/" + utils.URL2Filename(urlstr)
	outfile, err := os.Create(part_file)
	if err != nil {
		return err
	}
	defer outfile.Close()
	downloader := &Downloader{ReadCloser: body, percentage: a.percentage}
	_, err = io.Copy(outfile, downloader)
	return err
}

// 断点续传合并文件
func (a *Assassin) merge(urlstr string, part_size int) error {
	fii, err := os.OpenFile(strings.TrimRight(a.conf.Folder, "/")+"/"+utils.URL2Filename(urlstr), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	// 合并文件
	for i := 0; i < part_size; i++ {
		part_file := fmt.Sprintf("/tmp/%s.%d.part", utils.URL2Filename(urlstr), i)
		pf := &os.File{}
		pf, err = os.OpenFile(part_file, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		body := make([]byte, 0)
		body, err = ioutil.ReadAll(pf)
		if err != nil {
			return err
		}
		fii.Write(body)
		pf.Close()
	}
	return fii.Close()
}
