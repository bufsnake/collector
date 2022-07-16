package log

import (
	"fmt"
	"github.com/bufsnake/collector/pkg/utils"
	"strings"
	"sync"
	"time"
)

// ANSI属性控制码 https://www.cnblogs.com/cuckoo-/p/10741562.html
// ANSI属性控制码 https://icode.best/i/09251342647519
// https://github.com/wxnacy/study/blob/master/goland/src/progress/single.go
type Log struct {
	lock        *sync.Mutex
	percentages []*Percentage
	wait        bool // 当前是否已经进入等待状态
}

func NewLog() *Log {
	return &Log{percentages: make([]*Percentage, 0), lock: &sync.Mutex{}}
}

func (l *Log) Add(p *Percentage) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.percentages = append(l.percentages, p)
}

func (l *Log) Run() {
	for {
		if l.wait && l.IsFinish() {
			break
		}
		l.lock.Lock()
		percentages := make([]*Percentage, 0)
		for i := 0; i < len(l.percentages); i++ {
			if l.percentages[i].finish {
				continue
			}
			percentages = append(percentages, l.percentages[i])
		}
		for i := 0; i < len(percentages); i++ {
			percentages[i].println()
		}
		if len(percentages) != 0 {
			//fmt.Printf("\033[%dA\033[K\033[?25l", len(percentages)) // 向上移动line行，并清除对应行的内容
			fmt.Printf("\033[%dA\033[?25l", len(percentages)) // 向上移动line行
		}
		for i := 0; i < len(percentages); i++ {
			if !percentages[i].finish {
				continue
			}
			percentages[i].print(true)
		}
		l.lock.Unlock()
		time.Sleep(time.Second / 5)
	}
}

func (l *Log) IsFinish() bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.wait = true
	for i := 0; i < len(l.percentages); i++ {
		if !l.percentages[i].finish {
			return false
		}
	}
	fmt.Printf("\033[?25h")
	return true
}

func (p *Log) Println(a ...interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	fmt.Printf("\033[K\033[?25l") // 清除当前行
	fmt.Println(a...)
}

type Percentage struct {
	speed    int    // 网速
	progress int    // 进度
	total    int    // 总
	filename string // 文件名
	size     int    // 文件大小
	finish   bool   // 是否已经完成
	lock     *sync.Mutex
	single   bool // 是否为整体下载
}

func NewPercentage(filename string, total int, single bool) *Percentage {
	percentage := Percentage{progress: 0, filename: filename, total: total, lock: &sync.Mutex{}, single: single}
	go percentage.calcSpeed()
	return &percentage
}

func (p *Percentage) calcSpeed() {
	progress := 0
	for {
		p.lock.Lock()
		p.speed = p.progress - progress
		progress = p.progress
		p.lock.Unlock()
		time.Sleep(time.Second)
		if p.finish {
			break
		}
	}
}

func (p *Percentage) AddProgress(progress int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.progress += progress
	if p.progress >= p.total {
		p.progress = p.total
	}
}

func (p *Percentage) getFileName() string {
	filename := p.filename
	if len(filename) > 50 {
		filename = p.filename[:50]
	}
	return filename + strings.Repeat(" ", 50-len(filename))
}

func (p *Percentage) println() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.print(false)
	if p.progress != p.total {
		return
	}
	p.finish = true
}

func (p *Percentage) print(success bool) {
	if success {
		fmt.Printf("\033[K\033[?25l") // 清除当前行
	}
	filename := p.getFileName()
	progress := fmt.Sprintf("%.2f%%", (float64(p.progress)/float64(p.total))*100)
	if len(progress) <= 7 {
		progress = strings.Repeat(" ", 7-len(progress)) + progress
	}
	fav := "m"
	if p.single {
		fav = "s"
	}
	fmt.Printf("%s \033[K\033[?25l\n", fmt.Sprintf( // 清除当前行并打印进度
		"%s %s (%s)%s %s %s/s",
		filename,
		utils.ByteFormat(float64(p.total), utils.RawUnitB, true),
		fav,
		"progress:",
		progress,
		utils.ByteFormat(float64(p.speed), utils.RawUnitB, false),
	))
}
