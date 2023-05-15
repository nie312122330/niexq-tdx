package tdxlocal

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/nie312122330/niexq-gotools/dateext"
)

// 直接使用股票代码获取1分钟K线的存储位置
// kname  lday |minline
func Code2FilePath(vipdocDir, stCode6, kname string) string {
	mktStr := "sz"
	if strings.HasPrefix(stCode6, "00") || strings.HasPrefix(stCode6, "30") {
		mktStr = "sz"
	} else if strings.HasPrefix(stCode6, "60") {
		mktStr = "sh"
	}
	sufix := "lc1"
	if "lday" == kname {
		sufix = "day"
	}
	return fmt.Sprintf("%s/%s/%s/%s%s.%s", vipdocDir, mktStr, kname, mktStr, stCode6, sufix)
}

// 通达信日期，和时间转化为GO日期
func lc1mDateTime2Str(date, time uint16) time.Time {
	year := int(math.Floor(float64(date)/2048)) + 2004
	month := int(math.Floor(math.Mod(float64(date), 2048) / 100))
	day := int(math.Mod(math.Mod(float64(date), 2048), 100))
	hour := int(math.Floor(float64(time) / 60))
	miniu := int(math.Mod(float64(time), 60))
	acDate := dateext.WithDate(year, month, day, hour, miniu, 0)
	return acDate.Time
}
