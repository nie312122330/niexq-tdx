package tdxlocal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"path/filepath"
	"time"

	"github.com/nie312122330/niexq-gotools/dateext"
)

func StartParseTcxLc1m(tdxVipDir string) {
	markets := []string{"sz", "sh"}
	for _, m := range markets {
		lc1mDataDir := fmt.Sprintf("%s/%s/%s", tdxVipDir, m, "minline")
		sockFiles, _ := ioutil.ReadDir(lc1mDataDir)
		for _, file := range sockFiles {
			fileName := file.Name()
			stockCode := fileName[2:8]
			if m == "sz" {
				stockCode = "0" + stockCode
			} else {
				stockCode = "1" + stockCode
			}
			stFilePath := filepath.Join(lc1mDataDir, fileName)
			log.Printf("开始处理股票【%s】,分时文件为：【%s】", stockCode, stFilePath)
			parseStockLc1m(stFilePath, stockCode)
		}
	}
}

//7位股票代码
func parseStockLc1m(filePath string, stockCode string) {
	//确定文件名称
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	dataLen := len(data)

	for i := 0; i < dataLen; i = i + 32 {
		lc1mData := data[i : i+32]
		buf := bytes.NewBuffer(lc1mData)
		vo := &Lc1mLineVo{}
		binary.Read(buf, binary.LittleEndian, vo)
		dataTime := lc1mDateTime2Str(vo.Date, vo.Time)
		//保存数据
		log.Printf("时间:%s\n", dataTime)
		//这里就是通达信的数据
	}
}

//通达信日期，和时间转化为GO日期
func lc1mDateTime2Str(date, time uint16) time.Time {
	year := int(math.Floor(float64(date)/2048)) + 2004
	month := int(math.Floor(math.Mod(float64(date), 2048) / 100))
	day := int(math.Mod(math.Mod(float64(date), 2048), 100))
	hour := int(math.Floor(float64(time) / 60))
	miniu := int(math.Mod(float64(time), 60))
	acDate := dateext.WithDate(year, month, day, hour, miniu, 0)
	return acDate.Time
}

//1分红数据的数据结构
type Lc1mLineVo struct {
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
