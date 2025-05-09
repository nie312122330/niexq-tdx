package test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/axgle/mahonia"
	"github.com/nie312122330/niexq-gotools/dateext"
	"github.com/nie312122330/niexq-gotools/fileext"
	"github.com/nie312122330/niexq-tdx/tdx"
	tdxext "github.com/nie312122330/niexq-tdx/tdx-ext"
	tdxlocal "github.com/nie312122330/niexq-tdx/tdx-local"
	"github.com/shopspring/decimal"
)

var tdxConn *tdx.TdxConn

func init() {
	serverAddr := "120.76.1.198:7709"
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
		fmt.Printf("%v", v)
	}
}

// 测试  读取本地的1分钟线
func TestLocalFile(t *testing.T) {
	//K线数据测试,计算量与价的关系
	start := dateext.WithDate(2022, 9, 15, 9, 0, 0).Time
	end := dateext.WithDate(2022, 10, 10, 18, 0, 0).Time
	buy, sall, min, err := tdxlocal.Lc1mBarVoByTimeBuyMoney(start, end, `C:\zd_zsone\vipdoc`, "603296")
	fmt.Printf("买:%0.2f,卖:%0.2f,平:%0.2f,%v", buy, sall, min, err)
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
		slog.Info(fmt.Sprintf("%v:%v:%v %v  %v  %v ", v.Hour, v.Minus, v.Second, v.Price, v.Vol, v.Buyorsell))
	}
	slog.Info(fmt.Sprintf("历史分时成交返回数据【%d】条", len(resp.Datas)))
}

// 测试  今日分时行情
func TestQueryTodayFshq(t *testing.T) {
	res2, _, _ := tdxConn.QueryTodayFshq(0, "002008", 20221111, 2594)
	for idx, v := range res2.Datas {
		if idx <= 10 || (idx >= 110 && idx <= 130) || idx >= 230 {
			slog.Info(fmt.Sprintf("%v,%v,%v,%v,%v", v.DateTime.ToStr(), v.Price, v.AvgPrice, v.Vol, v.VolFlag))
		}
	}
}

// 测试  历史分时行情
func TestQueryFshq(t *testing.T) {
	res2, _, _ := tdxConn.QueryLsFshq(20221111, 1, "600519")
	for idx, v := range res2.Datas {
		if idx <= 10 || (idx >= 110 && idx <= 130) || idx >= 230 {
			slog.Info(fmt.Sprintf("%v,%v,%v,%v,%v", v.DateTime.ToStr(), v.Price, v.AvgPrice, v.Vol, v.VolFlag))
		}
	}
}

// 测试 集合竞价
func TestQueryJhjj(t *testing.T) {
	res3, _ := tdxConn.QueryTodayJhjj(1, "600322")
	slog.Info(fmt.Sprintf("集合竞价返回数据【%d】条", len(res3.Datas)))
	for _, v := range res3.Datas {
		slog.Info(fmt.Sprintf("%v", v))
	}
}

// 测试 查询1分钟的K线图
func TestQueryBarK1m(t *testing.T) {
	res4, _ := tdxConn.QueryLsBarK1m(1, "600322", 0, 100)
	slog.Info(fmt.Sprintf("1分钟的K线图返回数据【%d】条", len(res4.Datas)))
}

// 测试 查询指定最大量能及收盘价
func TestQueryDatesMaxVolAndClosePrice(t *testing.T) {
	maxVol := tdxext.QueryDateMaxVol(tdxConn, 20221118, 0, "002197")
	slog.Info(fmt.Sprintf("【%s】最大量为:%d", "002073", maxVol))
}

// 测试 查询股票名称
func TestQueryStName(t *testing.T) {
	name, _ := tdx.QueryStName(0, "000630", 3)
	slog.Info(name)
}

// 测试 查询历史的所有分时成交
func TestQueryLsFscj(t *testing.T) {
	datas := tdxext.QueryLsFscj(tdxConn, 20221111, 1, "600879")
	fmt.Println(len(datas))
	for _, v := range datas {
		fmt.Printf("%v", v)
	}
}

// 测试  1
func Test1(t *testing.T) {
	dateStr, _ := dateext.Now().Format("yyyyMMdd", false)
	dateInt := tdx.StrInt2Int(dateStr)
	datas, _ := tdxext.QueryFsHqAndMoney(tdxConn, int32(dateInt), 0, "002197", 20*10000)
	for _, v := range datas {
		fmt.Printf("%v %v %v %v %v %v", v.DateTime.ToStr(), v.Price, v.Vol, v.BigInMoney, v.BigOutMoney, v.BigMoneyCount)
	}
}

// 获取今日分时行情与交易金额的关系--- bigMoney元
func TestQueryTodayFsHqAndMoney(t *testing.T) {
	datas, _ := tdxext.QueryTodayFsHqAndMoney(tdxConn, 20221114, 0, "002397", 20*10000, 1105)
	for _, v := range datas {
		fmt.Printf("%v %v %v %v %v %v", v.DateTime.ToStr(), v.Price, v.Vol, v.BigInMoney, v.BigOutMoney, v.BigMoneyCount)
	}
}

func TestZtPrice(t *testing.T) {
	tdx.ZtPrice(1275, 0.1)
}

func TestAA(t *testing.T) {
	datas, _ := tdxext.QueryFsHqAndMoney(tdxConn, 20221102, 0, "002826", 200000)
	fmt.Printf("%v %v %v", datas[0].DateTime.ToStr(), datas[0].Price, datas[0].BigMoneyCount)
	fmt.Printf("%v %v %v", datas[len(datas)-1].DateTime.ToStr(), datas[len(datas)-1].Price, datas[len(datas)-1].BigMoneyCount)
	b, s, c := tdxext.CountDateLsFscj(tdxConn, 20221102, 0, "002826", 200000)
	fmt.Println(b, s, c)
}

func TestB1(t *testing.T) {

	// a := float32(9.4) * float32(100)
	// ss := decimal.NewFromFloat32(a)
	// s := ss.Floor()
	fmt.Println(decimal.NewFromFloat(5.460).Round(1).String())
	fmt.Println(decimal.NewFromFloat(5.450).Round(1).String())
	fmt.Println(decimal.NewFromFloat(5.449).Round(1).String())
	fmt.Println(decimal.NewFromFloat(5.4).Round(0).String())
	fmt.Println(decimal.NewFromFloat(5.5).Round(0).String())
	fmt.Println(decimal.NewFromFloat(5.6).Round(0).String())
}
