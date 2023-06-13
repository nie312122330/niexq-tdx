package test

import (
	"fmt"
	"testing"

	tdxlocal "github.com/nie312122330/niexq-tdx/tdx-local"
)

// 测试  读取股票列表
func TestReadDay(t *testing.T) {
	path := tdxlocal.Code2FilePath(`C:/zd_zsone/vipdoc`, "603112", "lday")
	vos := tdxlocal.ParseStockLc1dFile(path)
	tdxlocal.SortDayVosByDate(vos, false)
	//降序
	for _, v := range vos {
		fmt.Printf("%v  %f\n", v.Date, v.Amount)
	}
}

// 测试  计算涨停数量
func TestCountDayK10cm(t *testing.T) {

	nukm := tdxlocal.CountDayK10cm(`C:/zd_zsone/vipdoc`, "603335")
	fmt.Printf("数量%v\n", nukm)
}
