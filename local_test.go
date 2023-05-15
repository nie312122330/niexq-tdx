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
