package runner

import (
	"github.com/bufsnake/collector/config"
	"github.com/bufsnake/collector/pkg/assassin"
	"github.com/bufsnake/collector/pkg/log"
	"github.com/bufsnake/collector/pkg/utils"
	"strings"
	"sync"
	"time"
)

type runner struct {
	conf *config.Config
	log  *log.Log
}

func NewRunner(conf *config.Config) *runner {
	return &runner{conf: conf, log: log.NewLog()}
}

func (r *runner) Run(urlstrs []string) error {
	err := utils.CreateFolder(r.conf.Folder)
	if err != nil {
		return err
	}
	go func() {
		r.log.Run()
	}()
	wait := sync.WaitGroup{}
	urlstrc := make(chan string)
	for i := 0; i < r.conf.Thread; i++ {
		wait.Add(1)
		go r.down(&wait, r.conf, urlstrc)
	}
	for i := 0; i < len(urlstrs); i++ {
		urlstrc <- urlstrs[i]
	}
	close(urlstrc)
	wait.Wait()
	for !r.log.IsFinish() {
	}
	return nil
}

func (r *runner) down(dw *sync.WaitGroup, conf *config.Config, urlstrs chan string) {
	defer dw.Done()
	for urlstr := range urlstrs {
		asin := assassin.NewAssassin(conf, r.log)
		detect_retry := -1
	detect_again:
		dbc, chunked, cl, err := asin.DetectBreakpointContinuingly(urlstr)
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "content-length") {
				r.log.Println("detect breakpoint continuingly", err)
				continue
			}
			detect_retry++
			if detect_retry == r.conf.MaxRetry {
				continue
			}
			time.Sleep(2 * time.Second)
			goto detect_again
		}
		if chunked {
			retry := -1
		chunked_again:
			err = asin.ChunkedDownload(urlstr)
			if err != nil {
				if retry < r.conf.MaxRetry {
					retry++
					time.Sleep(2 * time.Second)
					goto chunked_again
				}
				r.log.Println("chunked download", err)
			}
			continue
		}
		if cl == 0 || cl == -1 {
			r.log.Println(urlstr, "content length error", cl)
			continue
		}
		if !dbc || strings.Contains(strings.ToLower(urlstr), "heapdump") {
			// 不支持断点续传
			retry := -1
		again:
			err = asin.OverallDownload(urlstr, int64(cl))
			if err != nil {
				if retry < r.conf.MaxRetry {
					retry++
					time.Sleep(2 * time.Second)
					goto again
				}
				r.log.Println("overall download", err)
			}
			continue
		}
		err = asin.BreakpointContinuinglyDownload(urlstr, cl)
		if err != nil {
			r.log.Println("breakpoint continuingly download", err)
		}
	}
}
