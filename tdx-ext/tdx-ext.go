package tdxext

import (
	"log"
	"sort"
	"strings"

	"github.com/nie312122330/niexq-tdx/tdx"
)

// 获取股票列表【主要用于今日行情的基础数据】
func QueryTodayStcokList(tdxConn *tdx.TdxConn) []tdx.StListItemVo {
	allVos := []tdx.StListItemVo{}
	//深证
	mkt := uint16(0)
	start := uint16(0)
	for {
		vos := tdxConn.QueryTodayStList(mkt, start)
		for _, v := range vos {
			if strings.HasPrefix(v.StCode, "00") || strings.HasPrefix(v.StCode, "30") {
				allVos = append(allVos, v)
			}
		}
		start += uint16(len(vos))
		if len(vos) <= 0 {
			break
		}
	}

	//上证
	mkt = 1
	start = 0
	for {
		vos := tdxConn.QueryTodayStList(mkt, start)
		for _, v := range vos {
			if strings.HasPrefix(v.StCode, "60") || strings.HasPrefix(v.StCode, "68") {
				allVos = append(allVos, v)
			}
		}
		start += uint16(len(vos))
		if len(vos) <= 0 {
			break
		}
	}
	return allVos
}

// 查询指定日期内的收盘价及分钟最大量
func QueryDatesMaxVolAndClosePrice(tdxConn *tdx.TdxConn, dates []int32, mkt byte, stCode string) (closePrice, maxVol int) {
	//排序
	tdx.BubbleSort(&dates)
	datas := []tdx.TdxFshqVo{}
	for _, v := range dates {
		res, _, _ := tdxConn.QueryLsFshq(v, mkt, stCode)
		datas = append(datas, res.Datas...)
	}
	closePrice = datas[len(datas)-1].Price
	for _, v := range datas {
		if v.Vol > maxVol {
			maxVol = v.Vol
		}
	}
	return closePrice, maxVol
}

// 查询今日分时成交[所有记录]
func QueryTodayFscj(tdxConn *tdx.TdxConn, mkt int16, stCode string) []tdx.TdxFscjVo {
	vos := []tdx.TdxFscjVo{}
	innerQueryTodayFscj(tdxConn, &vos, mkt, stCode, 0)
	sort.SliceStable(vos, func(i int, j int) bool {
		vo0 := vos[i]
		vo1 := vos[j]
		return vo0.Hour*3600+vo0.Minus*60+vo0.Second < vo1.Hour*3600+vo1.Minus*60+vo1.Second
	})
	return vos
}

// 查询历史所有的分时成交[所有记录]
func QueryLsFscj(tdxConn *tdx.TdxConn, date int32, mkt int16, stCode string) []tdx.TdxFscjVo {
	vos := []tdx.TdxFscjVo{}
	innerQueryLsFscj(tdxConn, &vos, date, mkt, stCode, 0)
	sort.SliceStable(vos, func(i int, j int) bool {
		vo0 := vos[i]
		vo1 := vos[j]
		return vo0.Hour*3600+vo0.Minus*60+vo0.Second < vo1.Hour*3600+vo1.Minus*60+vo1.Second
	})
	return vos
}

// 内部方法，查询今日分时成交【循环组装分页数据】
func innerQueryTodayFscj(tdxConn *tdx.TdxConn, vos *[]tdx.TdxFscjVo, mkt int16, stCode string, start int16) {
	resp, err := tdxConn.QueryTodayPageFscj(mkt, stCode, start, 1000)
	if nil != err {
		log.Printf("【%s】查询今日分时成交报错,%v", stCode, err)
		return
	}
	if len(resp.Datas) <= 0 {
		return
	}
	if len(resp.Datas) < 1000 {
		*vos = append(*vos, resp.Datas...)
		return
	}
	*vos = append(*vos, resp.Datas...)
	innerQueryTodayFscj(tdxConn, vos, mkt, stCode, start+1000)
}

// 内部方法，查询历史分时成交【循环组装分页数据】
func innerQueryLsFscj(tdxConn *tdx.TdxConn, vos *[]tdx.TdxFscjVo, date int32, mkt int16, stCode string, start int16) {
	resp, err := tdxConn.QueryLsPageFscj(date, mkt, stCode, start, 1000)
	if nil != err {
		log.Printf("【%s】查询今日分时成交报错,%v", stCode, err)
		return
	}
	if len(resp.Datas) <= 0 {
		return
	}
	if len(resp.Datas) < 1000 {
		*vos = append(*vos, resp.Datas...)
		return
	}
	*vos = append(*vos, resp.Datas...)
	innerQueryLsFscj(tdxConn, vos, date, mkt, stCode, start+1000)
}
