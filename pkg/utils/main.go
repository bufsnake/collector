package utils

import (
	"fmt"
	"os"
	"strings"
)

// b -> kb -> mb -> gb -> tb -> pb -> eb
const (
	RawUnitB  = " B"
	RawUnitKB = "KB"
	RawUnitMB = "MB"
	RawUnitGB = "GB"
	RawUnitTB = "TB"
	RawUnitPB = "PB"
	RawUnitEB = "EB"
)

// 1024.00 XB
func ByteFormat(data float64, unit string, format bool) string {
	newdata := data / 1024.0
	if newdata > 1 && unit != RawUnitEB {
		switch unit {
		case RawUnitB:
			return ByteFormat(newdata, RawUnitKB, format)
		case RawUnitKB:
			return ByteFormat(newdata, RawUnitMB, format)
		case RawUnitMB:
			return ByteFormat(newdata, RawUnitGB, format)
		case RawUnitGB:
			return ByteFormat(newdata, RawUnitTB, format)
		case RawUnitTB:
			return ByteFormat(newdata, RawUnitPB, format)
		case RawUnitPB:
			return ByteFormat(newdata, RawUnitEB, format)
		}
	}
	if format {
		padding_length := 7 - len(fmt.Sprintf("%.2f", data))
		if padding_length < 0 {
			padding_length = 0
		}
		return fmt.Sprintf(
			"%.2f%s",
			data,
			strings.Repeat(" ", padding_length)+unit,
		)
	}
	return fmt.Sprintf("%.2f%s", data, unit)
}

// URL转成文件名
func URL2Filename(urlstr string) string {
	urlstr = strings.ReplaceAll(urlstr, "://", "_")
	urlstr = strings.ReplaceAll(urlstr, ":", "-")
	return strings.ReplaceAll(urlstr, "/", "~")
}

// 创建文件夹
func CreateFolder(folder string) error {
	if existFolder(folder) {
		return nil
	}
	return os.MkdirAll(folder, os.ModePerm)
}

func existFolder(folder string) bool {
	_, err := os.Stat(folder) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
