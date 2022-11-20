package tdxext

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/nie312122330/niexq-gotools/dateext"
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

// 统计历史分时成交数据[所有记录]
func CountDateLsFscj(tdxConn *tdx.TdxConn, date int32, mkt int16, stCode string, bigMoney int) (b, s, c int64) {
	vos := QueryLsFscj(tdxConn, date, mkt, stCode)
	for _, v := range vos {
		//0 买，1-卖,2-竞价或平盘买入
		//涨停价需要看成主动性买单
		if v.Price == tdx.ZtPrice(v.PreClose, 0.1) {
			b += int64(v.Price * v.Vol)
		} else {
			if v.Price*v.Vol >= bigMoney {
				if v.Buyorsell == 0 {
					b += int64(v.Price * v.Vol)
				} else {
					s += int64(v.Price * v.Vol)
				}
			}
		}
	}
	return b, s, b - s
}

// 统计今日分时成交数据[所有记录]
func CountDateTodayFscj(tdxConn *tdx.TdxConn, mkt int16, stCode string, preClosePrice int, bigMoney int) (b, s, c int64) {
	vos := QueryTodayFscj(tdxConn, mkt, stCode, preClosePrice)
	for _, v := range vos {
		//0 买，1-卖,2-竞价或平盘买入
		//涨停价需要看成主动性买单
		if v.Price == tdx.ZtPrice(v.PreClose, 0.1) {
			b += int64(v.Price * v.Vol)
		} else {
			if v.Price*v.Vol >= bigMoney {
				if v.Buyorsell == 0 {
					b += int64(v.Price * v.Vol)
				} else {
					s += int64(v.Price * v.Vol)
				}
			}
		}
	}
	return b, s, b - s
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

// 查询历史所有的分时成交[所有记录]
func QueryTodayFscj(tdxConn *tdx.TdxConn, mkt int16, stCode string, preClosePrice int) []tdx.TdxFscjVo {
	vos := []tdx.TdxFscjVo{}
	innerQueryTodayFscj(tdxConn, &vos, mkt, stCode, preClosePrice, 0)
	sort.SliceStable(vos, func(i int, j int) bool {
		vo0 := vos[i]
		vo1 := vos[j]
		return vo0.Hour*3600+vo0.Minus*60+vo0.Second < vo1.Hour*3600+vo1.Minus*60+vo1.Second
	})
	return vos
}

// 获取今日分时行情与交易金额的关系--- bigMoney元
func QueryTodayFsHqAndMoney(tdxConn *tdx.TdxConn, todayInt int32, mkt int16, stCode string, bigMoney int, preClosePrice int) (resVos []TdxExtTodayMoney, err error) {
	fscjVos := QueryTodayFscj(tdxConn, mkt, stCode, preClosePrice)
	if len(fscjVos) <= 0 {
		return resVos, fmt.Errorf("未获取到分时成交")
	}
	resp, preClosePrice, err := tdxConn.QueryTodayFshq(byte(mkt), stCode, int(todayInt), preClosePrice)
	if nil != err {
		return resVos, err
	}
	return concatFsHqAndMoney(fscjVos, resp.Datas, bigMoney, preClosePrice)
}

// 获取分时行情与交易金额的关系--- bigMoney元
func QueryFsHqAndMoney(tdxConn *tdx.TdxConn, date int32, mkt int16, stCode string, bigMoney int) (resVos []TdxExtTodayMoney, err error) {
	fscjVos := QueryLsFscj(tdxConn, date, mkt, stCode)
	if len(fscjVos) <= 0 {
		return resVos, fmt.Errorf("未获取到分时成交")
	}
	resp, preClosePrice, err := tdxConn.QueryLsFshq(date, byte(mkt), stCode)
	if nil != err {
		return resVos, err
	}
	return concatFsHqAndMoney(fscjVos, resp.Datas, bigMoney, preClosePrice)
}

func concatFsHqAndMoney(fscjVos []tdx.TdxFscjVo, fshqVos []tdx.TdxFshqVo, bigMoney int, preClosePrice int) (resVos []TdxExtTodayMoney, err error) {
	//增加分时行情  的竞价金额
	hq01Time := time.Time(fshqVos[0].DateTime)
	if fscjVos[0].Hour == 9 && fscjVos[0].Minus == 25 {
		tmpHqVos := []tdx.TdxFshqVo{}
		dateTime := dateext.WithDate(hq01Time.Year(), int(hq01Time.Month()), hq01Time.Day(), 9, 25, 0)
		tmpHqVos = append(tmpHqVos, tdx.TdxFshqVo{
			DateTime: tdx.TdxJsonTime(dateTime.Time),
			Price:    fscjVos[0].Price,
			AvgPrice: fscjVos[0].Price,
			Vol:      fscjVos[0].Vol,
			VolFlag:  2,
		})
		tmpHqVos = append(tmpHqVos, fshqVos...)
		fshqVos = tmpHqVos
	}

	if fscjVos[len(fscjVos)-1].Hour == 15 && fscjVos[len(fscjVos)-1].Minus == 0 {
		dateTime := dateext.WithDate(hq01Time.Year(), int(hq01Time.Month()), hq01Time.Day(), 15, 0, 0)
		fshqVos = append(fshqVos, tdx.TdxFshqVo{
			DateTime: tdx.TdxJsonTime(dateTime.Time),
			Price:    fscjVos[len(fscjVos)-1].Price,
			AvgPrice: fscjVos[len(fscjVos)-1].Price,
			Vol:      fscjVos[len(fscjVos)-1].Vol,
			VolFlag:  2,
		})
	}

	//按时间分组 成交数据
	fscjMaps := make(map[int]*[]tdx.TdxFscjVo)
	for _, v := range fscjVos {
		key := v.Hour*60 + v.Minus
		val, ok := fscjMaps[key]
		if ok {
			*val = append(*val, v)
		} else {
			fscjMaps[key] = &[]tdx.TdxFscjVo{v}
		}
	}

	getBigInfo := func(time int) (b, s, c int64) {
		fscjTimeDatas, ok := fscjMaps[time]
		if !ok {
			return b, s, c
		}
		//0 买，1-卖,2-竞价或平盘买入
		for _, v := range *fscjTimeDatas {
			//涨停价需要看成主动性买单
			if v.Price == tdx.ZtPrice(v.PreClose, 0.1) {
				b += int64(v.Price * v.Vol)
			} else {
				if v.Price*v.Vol >= bigMoney {
					if v.Buyorsell == 0 {
						b += int64(v.Price * v.Vol)
					} else {
						s += int64(v.Price * v.Vol)
					}
				}
			}
		}
		return b, s, b - s
	}
	moneyCount := int64(0)
	for _, vo := range fshqVos {
		voTime := time.Time(vo.DateTime)
		//计算买单，卖单
		voTimtInt := voTime.Hour()*60 + voTime.Minute()
		// fmt.Println(voTimtInt)
		b, s, c := getBigInfo(voTimtInt)
		moneyCount += c
		//获取这一分钟的 大单
		resVos = append(resVos, TdxExtTodayMoney{
			TdxFshqVo:     vo,
			PreClosePrice: preClosePrice,
			BigInMoney:    b,
			BigOutMoney:   s,
			BigMoneyCount: moneyCount,
		})
	}
	return resVos, nil
}

// 内部方法，查询历史分时成交【循环组装分页数据】
func innerQueryLsFscj(tdxConn *tdx.TdxConn, vos *[]tdx.TdxFscjVo, date int32, mkt int16, stCode string, start int16) {
	resp, err := tdxConn.QueryLsPageFscj(date, mkt, stCode, start, 1000)
	if nil != err {
		log.Printf("【%s】查询历史分时成交报错,%v", stCode, err)
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

// 内部方法，查询历史分时成交【循环组装分页数据】
func innerQueryTodayFscj(tdxConn *tdx.TdxConn, vos *[]tdx.TdxFscjVo, mkt int16, stCode string, preClosePrice int, start int16) {
	resp, err := tdxConn.QueryTodayPageFscj(mkt, stCode, preClosePrice, start, int16(1000))
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
	innerQueryTodayFscj(tdxConn, vos, mkt, stCode, preClosePrice, int16(start+1000))
}
