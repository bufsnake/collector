package assassin

import (
	"github.com/bufsnake/collector/pkg/log"
	"io"
	"math/rand"
)

// 实现 ReadCloser 可获取下载进度
type Downloader struct {
	io.ReadCloser
	percentage *log.Percentage
	chunked    bool
	count      int
}

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.ReadCloser.Read(p)
	if d.chunked {
		intn := rand.Intn(n+2) + 100
		d.percentage.AddTotal(n + intn)
		d.count += intn
	} else {
		d.count += n
	}
	d.percentage.AddProgress(n)
	return
}
