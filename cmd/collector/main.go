package main

import (
	"flag"
	"fmt"
	"github.com/bufsnake/collector/config"
	"github.com/bufsnake/collector/internal/runner"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	conf := config.Config{
		Timeout:  10,
		Folder:   "test",
		Thread:   10,
		MaxRetry: 5,
	}
	flag.StringVar(&conf.Target, "url", "", "specify URL to download")
	flag.StringVar(&conf.TargetList, "url-file", "", "get URL from specified file for batch download")
	flag.StringVar(&conf.Folder, "folder", "./files", "file save location")
	flag.IntVar(&conf.Thread, "thread", 10, "batch write on and chunked download threads")
	flag.IntVar(&conf.MaxRetry, "max-retry", 10, "after a failed download, the maximum number of attempts")
	flag.IntVar(&conf.Timeout, "timeout", 10, "the connection times out, and the download time defaults to 2 hours")
	flag.Parse()
	urlstrs := make([]string, 0)
	if conf.Target != "" {
		urlstrs = append(urlstrs, conf.Target)
	} else if conf.TargetList != "" {
		probes := make(map[string]bool)
		file, err := os.ReadFile(conf.TargetList)
		if err != nil {
			fmt.Println(err)
			return
		}
		urls := strings.Split(string(file), "\n")
		for i := 0; i < len(urls); i++ {
			urls[i] = strings.Trim(urls[i], " \t\r")
			if urls[i] == "" {
				continue
			}
			urls[i] = strings.TrimSpace(urls[i])
			if _, ok := probes[urls[i]]; ok {
				continue
			}
			probes[urls[i]] = true
			urlstrs = append(urlstrs, urls[i])
		}
	} else {
		flag.Usage()
		return
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		<-c
		fmt.Printf("\033[?25h") // 当ctrl+c时显示光标
		os.Exit(0)
	}()
	err := runner.NewRunner(&conf).Run(urlstrs)
	if err != nil {
		fmt.Println(err)
		return
	}
}
