package test

import (
	"fmt"
	"testing"

	"github.com/nie312122330/niexq-gotools/jsonext"
	tdxext "github.com/nie312122330/niexq-tdx/tdx-ext"
)

// 测试  获取A股节假日
func TestQueryHolidys(t *testing.T) {
	datas, _ := tdxext.TdxHolidys()
	fmt.Println(jsonext.ToStrOk(len(datas)))
}
