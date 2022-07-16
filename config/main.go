package config

type Config struct {
	Target     string // 下载单个文件
	TargetList string // 下载多个文件
	Folder     string // 目标文件夹，默认为当前文件夹的files文件夹
	Thread     int    // 多线程下载
	MaxRetry   int    // 失败重试次数
	Timeout    int    // 超时
}
