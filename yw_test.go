package test

import (
	"log"
	"testing"

	"github.com/nie312122330/niexq-tdx/tdx"
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

//测试  分时成交
func TestQueryFscj(t *testing.T) {
	res, _ := tdxConn.QueryFscj(0, "000630", 0, 100)
	log.Printf("分时成交返回数据【%d】条\r\n", len(res.Datas))
}

//测试  分时行情
func TestQueryFshq(t *testing.T) {
	res2, _ := tdxConn.QueryFshq(20220711, 0, "000630")
	log.Printf("分时行情返回数据【%d】条\n", len(res2.Datas))
}

//测试 集合竞价
func TestQueryJhjj(t *testing.T) {
	res3, _ := tdxConn.QueryJhjj(1, "600021")
	log.Printf("集合竞价返回数据【%d】条\n", len(res3.Datas))
}

//测试 查询1分钟的K线图
func TestQueryBarK1m(t *testing.T) {
	res4, _ := tdxConn.QueryBarK1m(1, "600021", 0, 100)
	log.Printf("1分钟的K线图返回数据【%d】条\n", len(res4.Datas))
}

//测试 查询指定日志的最大量能及收盘价
func TestQueryDatesMaxVolAndClosePrice(t *testing.T) {
	dates := []int32{20220712, 20220711}
	closePrice, maxVol := tdxConn.QueryDatesMaxVolAndClosePrice(dates, 0, "002073")
	log.Printf("【%s】最大量为:%d,最大金额:%d\n", "002073", maxVol, closePrice)
}

func TestQueryStName(t *testing.T) {
	name, _ := tdx.QueryStName(0, "000630", 3)
	log.Printf("%s\n", name)
}
