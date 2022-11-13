package test

import (
	"fmt"
	"log"
	"testing"

	"github.com/axgle/mahonia"
	"github.com/nie312122330/niexq-gotools/dateext"
	"github.com/nie312122330/niexq-gotools/fileext"
	"github.com/nie312122330/niexq-tdx/tdx"
	tdxext "github.com/nie312122330/niexq-tdx/tdx-ext"
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

// 测试  读取股票列表
func TestQueryStList(t *testing.T) {
	datas := tdxext.QueryTodayStcokList(tdxConn)
	fmt.Println(len(datas))
	for _, v := range datas {
		fmt.Printf("%v\n", v)
	}
}

// 测试  读取本地的1分钟线
func TestLocalFile(t *testing.T) {
	//K线数据测试,计算量与价的关系
	start := dateext.WithDate(2022, 9, 15, 9, 0, 0).Time
	end := dateext.WithDate(2022, 10, 10, 18, 0, 0).Time
	buy, sall, min := tdxlocal.Lc1mBarVoByTimeBuyMoney(start, end, `C:\zd_zsone\vipdoc`, "600171")
	fmt.Printf("买:%0.2f,卖:%0.2f,平:%0.2f\n", buy, sall, min)
}

// 测试  字符集合转换
func TestGBK2UTF8(t *testing.T) {
	readDatas, _ := fileext.ReadFileByte("test_gbk.txt")
	enc := mahonia.NewDecoder("GBK")
	result := enc.ConvertString(string(readDatas))
	fmt.Println(result)
}

// 测试  历史分时成交
func TestQueryLsPageFscj(t *testing.T) {
	resp, _ := tdxConn.QueryLsPageFscj(20221111, 0, "002197", 0, 1000)
	for _, v := range resp.Datas {
		log.Printf("%v:%v:%v %v  %v  %v \n", v.Hour, v.Minus, v.Second, v.Price, v.Vol, v.Buyorsell)
	}
	log.Printf("历史分时成交返回数据【%d】条\r\n", len(resp.Datas))
}

// 测试  今日分时行情
func TestQueryTodayFshq(t *testing.T) {
	res2, _, _ := tdxConn.QueryTodayFshq(0, "002008", 20221111, 2594)
	for idx, v := range res2.Datas {
		if idx <= 10 || (idx >= 110 && idx <= 130) || idx >= 230 {
			log.Printf("%v,%v,%v,%v,%v\n", v.DateTime.ToStr(), v.Price, v.AvgPrice, v.Vol, v.VolFlag)
		}
	}
}

// 测试  历史分时行情
func TestQueryFshq(t *testing.T) {
	res2, _, _ := tdxConn.QueryLsFshq(20221111, 1, "600519")
	for idx, v := range res2.Datas {
		if idx <= 10 || (idx >= 110 && idx <= 130) || idx >= 230 {
			log.Printf("%v,%v,%v,%v,%v\n", v.DateTime.ToStr(), v.Price, v.AvgPrice, v.Vol, v.VolFlag)
		}
	}
}

// 测试 集合竞价
func TestQueryJhjj(t *testing.T) {
	res3, _ := tdxConn.QueryTodayJhjj(1, "600322")
	log.Printf("集合竞价返回数据【%d】条\n", len(res3.Datas))
	for _, v := range res3.Datas {
		log.Printf("%v\n", v)
	}
}

// 测试 查询1分钟的K线图
func TestQueryBarK1m(t *testing.T) {
	res4, _ := tdxConn.QueryLsBarK1m(1, "600322", 0, 100)
	log.Printf("1分钟的K线图返回数据【%d】条\n", len(res4.Datas))
}

// 测试 查询指定最大量能及收盘价
func TestQueryDatesMaxVolAndClosePrice(t *testing.T) {
	dates := []int32{20220929}
	closePrice, maxVol := tdxext.QueryDatesMaxVolAndClosePrice(tdxConn, dates, 1, "600322")
	log.Printf("【%s】最大量为:%d,最大金额:%d\n", "002073", maxVol, closePrice)
	res3, _ := tdxConn.QueryTodayJhjj(1, "600322")
	dv := 0
	for _, v := range res3.Datas {
		if v.Hour == 9 {
			dv = v.Vol
		}
	}
	log.Printf("集合竞价Vol【%d】\n", dv)
}

// 测试 查询股票名称
func TestQueryStName(t *testing.T) {
	name, _ := tdx.QueryStName(0, "000630", 3)
	log.Printf("%s\n", name)
}

// 测试 查询历史的所有分时成交
func TestQueryLsFscj(t *testing.T) {
	datas := tdxext.QueryLsFscj(tdxConn, 20221111, 1, "600879")
	fmt.Println(len(datas))
	for _, v := range datas {
		fmt.Printf("%v\n", v)
	}
}

// 测试  1
func Test1(t *testing.T) {
	dateStr, _ := dateext.Now().Format("yyyyMMdd", false)
	dateInt := tdx.StrInt2Int(dateStr)
	datas, _ := tdxext.QueryFsHqAndMoney(tdxConn, int32(dateInt), 0, "002197", 20*10000)
	for _, v := range datas {
		fmt.Printf("%v %v %v %v %v %v\n", v.DateTime.ToStr(), v.Price, v.Vol, v.BigInMoney, v.BigOutMoney, v.BigMoneyCount)
	}
}

func TestZtPrice(t *testing.T) {
	tdx.ZtPrice(1275, 0.1)
}
