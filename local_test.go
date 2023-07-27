package test

import (
	"fmt"
	"testing"

	tdxlocal "github.com/nie312122330/niexq-tdx/tdx-local"
)

// 测试  读取股票列表
func TestReadDay(t *testing.T) {
	path := tdxlocal.Code2FilePath(`C:/zd_zsone/vipdoc`, "603296", "lday")
	vos, _ := tdxlocal.ParseStockLc1dFile(path)
	tdxlocal.SortDayVosByDate(vos, false)
	//降序
	for _, v := range vos {
		fmt.Printf("%v  %f\n", v.Date, v.Amount)
	}
}

// 测试  计算涨停数量
func TestCountDayK10cm(t *testing.T) {

	nukm, _ := tdxlocal.CountDayK10cm(`C:/zd_zsone/vipdoc`, "603296")
	fmt.Printf("数量%v\n", nukm)
}
