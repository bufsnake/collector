package assassin

import (
	"github.com/bufsnake/collector/pkg/log"
	"io"
)

// 实现 ReadCloser 可获取下载进度
type Downloader struct {
	io.ReadCloser
	percentage *log.Percentage
}

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.ReadCloser.Read(p)
	d.percentage.AddProgress(n)
	return
}
