package tdxext

import "github.com/nie312122330/niexq-tdx/tdx"

// 两家
type TdxExtTodayMoney struct {
	tdx.TdxFshqVo
	PreClosePrice int `json:"preClosePrice"`
	BigInMoney    int `json:"bigInMoney"`
	BigOutMoney   int `json:"bigOutMoney"`
	BigMoneyCount int `json:"bigMoneyCount"`
}
