package test

import (
	"fmt"
	"log"
	"testing"

	"github.com/axgle/mahonia"
	"github.com/nie312122330/niexq-gotools/dateext"
	"github.com/nie312122330/niexq-gotools/fileext"
	"github.com/nie312122330/niexq-tdx/tdx"
	tdxlocal "github.com/nie312122330/niexq-tdx/tdx-local"
)

var tdxConn *tdx.TdxConn

func init() {
	serverAddr := "119.147.212.81:7709"
	tdxConn1, err := tdx.NewTdxConn("conn1", serverAddr)
	if nil != err {
		panic(err)
	}
	tdxConn = tdxConn1
}

func TestLocalFile(t *testing.T) {
	//K线数据测试,计算量与价的关系
	start := dateext.WithDate(2022, 9, 21, 9, 0, 0).Time
	end := dateext.WithDate(2022, 9, 21, 18, 0, 0).Time
	datas := tdxlocal.Lc1mBarVoByTime(start, end, `C:\zd_zsone\vipdoc`, "600456")
	//filePath := tdxlocal.Code2FilePath(`C:\zd_zsone\vipdoc`, "600456")
	// datas := tdxlocal.ParseStockLc1mFile(filePath)
	for _, v := range datas {
		fmt.Printf("%s:%v,%v\n", v.DateTime.String(), v.Amount, v.Blzd)
	}
}

// 测试  字符集合转换
func TestGBK2UTF8(t *testing.T) {
	readDatas, _ := fileext.ReadFileByte("test_gbk.txt")
	enc := mahonia.NewDecoder("GBK")
	result := enc.ConvertString(string(readDatas))
	fmt.Println(result)
}

// 测试  分时成交
func TestQueryFscj(t *testing.T) {
	res, _ := tdxConn.QueryFscj(1, "600322", 3200, 10000)
	log.Printf("分时成交返回数据【%d】条\r\n", len(res.Datas))
}

// 测试  分时行情
func TestQueryFshq(t *testing.T) {
	res2, pc, _ := tdxConn.QueryFshq(20221019, 1, "600322")

	log.Printf("分时行情昨日收盘[%d],长度[%d],第1条为[%v]\n", pc, len(res2.Datas), res2.Datas[0])
}

// 测试 集合竞价
func TestQueryJhjj(t *testing.T) {
	res3, _ := tdxConn.QueryJhjj(1, "600322")
	log.Printf("集合竞价返回数据【%d】条\n", len(res3.Datas))
}

// 测试 查询1分钟的K线图
func TestQueryBarK1m(t *testing.T) {
	res4, _ := tdxConn.QueryBarK1m(1, "600322", 0, 100)
	log.Printf("1分钟的K线图返回数据【%d】条\n", len(res4.Datas))
}

// 测试 查询指定日志的最大量能及收盘价
func TestQueryDatesMaxVolAndClosePrice(t *testing.T) {
	dates := []int32{20220929}
	closePrice, maxVol := tdxConn.QueryDatesMaxVolAndClosePrice(dates, 1, "600322")
	log.Printf("【%s】最大量为:%d,最大金额:%d\n", "002073", maxVol, closePrice)
	res3, _ := tdxConn.QueryJhjj(1, "600322")
	dv := 0
	for _, v := range res3.Datas {
		if v.Hour == 9 {
			dv = v.Vol
		}
	}
	log.Printf("集合竞价Vol【%d】\n", dv)
}

func TestQueryStName(t *testing.T) {
	name, _ := tdx.QueryStName(0, "000630", 3)
	log.Printf("%s\n", name)
}

func TestReadTdxExportTxtFile(t *testing.T) {
	result := tdx.ReadTdxExportTxtFile("C:/Users/niexq/Desktop/20220722.txt")
	log.Printf("%v\n", result)
}

func TestA(t *testing.T) {
	result := tdx.ReadTdxExportTxtFile("C:/Users/niexq/Desktop/20220722.txt")
	log.Printf("%v\n", result)
}

func TestA1(t *testing.T) {
	jhjjvol(1, "601965", 20220929)
}

func jhjjvol(mkt byte, stCode string, preDay int32) {
	_, yestodayMaxVol := tdxConn.QueryDatesMaxVolAndClosePrice([]int32{preDay}, mkt, stCode)
	res3, _ := tdxConn.QueryJhjj(int16(mkt), stCode)
	closePrice, openPrice, jhjjVOl := jjJyData(res3.Datas)
	rato := float32(jhjjVOl) / float32(yestodayMaxVol)
	manzu := ""
	if rato > 0.3 {
		manzu = "Y"
	} else {
		manzu = "N"
	}
	jieguo := "跌"
	if closePrice > openPrice {
		jieguo = "涨"
	} else if openPrice == closePrice {
		jieguo = "平"
	} else {
		jieguo = "跌"
	}
	fmt.Printf("%s,%s,%s,昨日[%d]单分钟最大:%d,今日竞价量:%d,今日开盘:%d,今日收盘%d,比率:%f\n", stCode, manzu, jieguo, preDay, yestodayMaxVol, jhjjVOl, openPrice, closePrice, rato)
}

func jjJyData(datas []tdx.TdxJhjjVo) (closePrice, openPrice, jhjjVOl int) {
	cp := datas[len(datas)-1].Price
	op := 0
	ov := 0
	for i := 0; i < len(datas); i++ {
		if datas[i].Hour == 9 {
			op = datas[i].Price
			ov = datas[i].Vol
		}
	}
	return cp, op, ov
}
