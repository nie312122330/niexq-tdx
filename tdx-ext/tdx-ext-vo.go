package tdxext

import "github.com/nie312122330/niexq-tdx/tdx"

// 两家
type TdxExtTodayMoney struct {
	tdx.TdxFshqVo
	PreClosePrice int   `json:"preClosePrice"`
	BigInMoney    int64 `json:"bigInMoney"`
	BigOutMoney   int64 `json:"bigOutMoney"`
	BigMoneyCount int64 `json:"bigMoneyCount"`
}
