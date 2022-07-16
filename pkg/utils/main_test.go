package utils

import (
	"fmt"
	"testing"
)

func TestURL2Filename(t *testing.T) {
	fmt.Println(URL2Filename("http://www.bufsnake.com:9120/web/web.zip"))
}
