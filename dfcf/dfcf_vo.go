package dfcf

import (
	"fmt"
	"time"
)

type DcK1mResultVo struct {
	Data struct {
		Trends []string `json:"trends"`
	} `json:"data"`
}

// 集合竞价视图对象
type DcJhjjVo struct {
	Hour   int `json:"Hour"`
	Minus  int `json:"minus"`
	Second int `json:"second"`
	Price  int `json:"price"`
	Vol    int `json:"vol"`
	UnVol  int `json:"unVol"`
	UnFlag int `json:"unFlag"`
}

// 分时成交视图对象
type DcFscjVo struct {
	Hour      int `json:"Hour"`
	Minus     int `json:"minus"`
	Second    int `json:"second"`
	Price     int `json:"price"`
	Vol       int `json:"vol"`
	Num       int `json:"num"`
	Buyorsell int `json:"buyorsell"`
}

// 分时行情
type DcFshqVo struct {
	DateTime   DcJsonTime `json:"dateTime"`
	Price      int        `json:"price"`
	Vol        int        `json:"vol"`
	VolFlag    int        `json:"volFlag"`
	UnKonwData int        `json:"unKonwData"` //不晓得这个是什么，也不知道解析对没
}

// 1分钟的线
type DcBarK1mVo struct {
	DateTime DcJsonTime `json:"dateTime"`
	Open     int        `json:"open"`
	Close    int        `json:"close"`
	High     int        `json:"high"`
	Low      int        `json:"low"`
	Vol      int64      `json:"vol"`
	Money    int64      `json:"money"`
}

// 东方财富 分时图实时监控
type DcK1mNotifyVo struct {
	StockCode string
	Datas     []string
}

// 东方财富 分时图实时监控
type DcK1mRespVo struct {
	Rc   int `json:"rc"`
	Data struct {
		Trends []string `json:"trends"`
	} `json:"data"`
}

var TIME_LAYOUT string = "2006-01-02 15:04:05"

type DcJsonTime time.Time

// 实现它的json序列化方法
func (t DcJsonTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format(TIME_LAYOUT))
	return []byte(stamp), nil
}

// 扩展ToStr方法
func (t DcJsonTime) ToStr() string {
	return time.Time(t).Format(TIME_LAYOUT)
}
