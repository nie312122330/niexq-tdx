package tdx

import (
	"fmt"
	"time"
)

var TIME_LAYOUT string = "2006-01-02 15:04:05"

// 响应
type TdxRespBaseVo[T any] struct {
	Market  int    `json:"market"`
	StCode  string `json:"stCode"`
	TdxFunc uint16 `json:"tdxFunc"`
	Datas   []T    `json:"datas"`
}

// 集合竞价视图对象
type TdxJhjjVo struct {
	Hour   int `json:"Hour"`
	Minus  int `json:"minus"`
	Second int `json:"second"`
	Price  int `json:"price"`
	Vol    int `json:"vol"`
	UnVol  int `json:"unVol"`
	UnFlag int `json:"unFlag"`
	UnData int `json:"unData"`
}

// 分时成交视图对象
type TdxFscjVo struct {
	Hour      int `json:"Hour"`
	Minus     int `json:"minus"`
	Second    int `json:"second"`
	Price     int `json:"price"`
	Vol       int `json:"vol"`
	Num       int `json:"num"`
	Buyorsell int `json:"buyorsell"`
}

// 分时行情
type TdxFshqVo struct {
	DateTime   TdxJsonTime `json:"dateTime"`
	Price      int         `json:"price"`
	Vol        int         `json:"vol"`
	VolFlag    int         `json:"volFlag"`
	UnKonwData int         `json:"unKonwData"` //不晓得这个是什么，也不知道解析对没
}

// 1分钟的线
type TdxBarK1mVo struct {
	DateTime TdxJsonTime `json:"dateTime"`
	Open     int         `json:"open"`
	Close    int         `json:"close"`
	High     int         `json:"high"`
	Low      int         `json:"low"`
	Vol      int64       `json:"vol"`
	Money    int64       `json:"money"`
}

// 股票列表
type StListItemVo struct {
	StCode   string `json:"stCode"`
	StName   string `json:"stName"`
	PreClose int    `json:"preClose"`
}

type TdxJsonTime time.Time

// 实现它的json序列化方法
func (t TdxJsonTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(t).Format(TIME_LAYOUT))
	return []byte(stamp), nil
}

// 扩展ToStr方法
func (t TdxJsonTime) ToStr() string {
	return time.Time(t).Format(TIME_LAYOUT)
}

// 1分钟的线
type TdxTxtStVo struct {
	StCode string `json:"stCode"`
	StMkt  int16  `json:"stMkt"`
	StName string `json:"stName"`
	StZf   int    `json:"stZf"`
	StXJ   int    `json:"stXj"`
	StZd   int    `json:"stZd"`
}
