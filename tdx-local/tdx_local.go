package tdxlocal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"math"
	"strings"
	"time"

	"github.com/nie312122330/niexq-gotools/dateext"
)

// 时间区间内的数据
func Lc1mBarVoByTime(start, end time.Time, vipdocDir, stCode string) []Lc1mBarVo {
	filePath := Code2FilePath(vipdocDir, stCode)
	srcDatas := ParseStockLc1mFile(filePath)
	vos := []Lc1mBarVo{}
	for _, v := range srcDatas {
		if v.DateTime.After(start) && v.DateTime.Before(end) {
			vos = append(vos, v)
		}
	}
	return vos
}

// 解析文件
func ParseStockLc1mFile(filePath string) []Lc1mBarVo {
	//确定文件名称
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	dataLen := len(data)
	vos := []Lc1mBarVo{}
	for i := 0; i < dataLen; i = i + 32 {
		lc1mData := data[i : i+32]
		buf := bytes.NewBuffer(lc1mData)
		data := &TdxLc1mData{}
		binary.Read(buf, binary.LittleEndian, data)
		dataTime := lc1mDateTime2Str(data.Date, data.Time)
		vo := Lc1mBarVo{
			DateTime: dataTime,
			Open:     data.Open,
			High:     data.High,
			Low:      data.Low,
			Close:    data.Close,
			Amount:   data.Amount,
			Qty:      data.Qty,
			Blzd:     data.Blzd,
		}
		vos = append(vos, vo)
	}
	return vos
}

// 直接使用股票代码获取1分钟K线的存储位置
func Code2FilePath(vipdocDir, stCode6 string) string {
	mktStr := "sz"
	if strings.HasPrefix(stCode6, "00") || strings.HasPrefix(stCode6, "30") {
		mktStr = "sz"
	} else if strings.HasPrefix(stCode6, "60") {
		mktStr = "sh"
	}
	return fmt.Sprintf("%s/%s/%s/%s%s.lc1", vipdocDir, mktStr, "minline", mktStr, stCode6)
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

// 1分红数据的数据结构
type TdxLc1mData struct {
	//  日期计算方式
	//	year := int(math.Floor(float64(nyr)/2048)) + 2004
	//	month := int(math.Floor(math.Mod(float64(nyr), 2048) / 100))
	// 	day := int(math.Mod(math.Mod(float64(nyr), 2048), 100))
	Date uint16 //日期
	//   时间计算方式
	//   hour := int(math.Floor(float64(sfm) / 60))
	//   miniu := int(math.Mod(float64(sfm), 60))
	Time   uint16  //时间
	Open   float32 //开盘价
	High   float32 //最高价
	Low    float32 //最低价
	Close  float32 //收盘价
	Amount float32 //总成交金额
	Qty    uint32  //  总成交量股数
	Blzd   uint32  //其他
}

type Lc1mBarVo struct {
	DateTime time.Time //时间日期
	Open     float32   //开盘价
	High     float32   //最高价
	Low      float32   //最低价
	Close    float32   //收盘价
	Amount   float32   //总成交金额
	Qty      uint32    //总成交量股数
	Blzd     uint32    //其他
}
