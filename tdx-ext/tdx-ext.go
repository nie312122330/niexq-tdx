package tdxext

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nie312122330/niexq-gotools/dateext"
	"github.com/nie312122330/niexq-tdx/tdx"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 获取股票列表【主要用于今日行情的基础数据】
func QueryTodayStcokList(tdxConn *tdx.TdxConn) []tdx.StListItemVo {
	allVos := []tdx.StListItemVo{}
	//深证
	mkt := uint16(0)
	start := uint16(0)
	for {
		vos := tdxConn.QueryTodayStList(mkt, start)
		slog.Info(fmt.Sprintf("获取到sz股票数量为:%d", len(vos)))
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
		slog.Info(fmt.Sprintf("获取到sh股票数量为:%d", len(vos)))
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

// 查询指定日期单分钟最大量--使用分时成交计算
func QueryDateMaxVol(tdxConn *tdx.TdxConn, dateInt int32, mkt byte, stCode string) (maxVol int64) {
	maxVol = int64(-1)
	//排序
	datas := QueryLsFscj(tdxConn, dateInt, int16(mkt), stCode)
	if len(datas) <= 0 {
		return maxVol
	}
	//按时间分组后的分时成交数据
	maxVolMap := make(map[int]int64)
	for _, v := range datas {
		key := v.Hour*60 + v.Minus
		val, ok := maxVolMap[key]
		if ok {
			maxVolMap[key] = val + int64(v.Vol)
		} else {
			maxVolMap[key] = int64(v.Vol)
		}
	}

	//获取最大值
	for _, v := range maxVolMap {
		if v > maxVol {
			maxVol = v
		}
	}

	return maxVol
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
		slog.Info(fmt.Sprintf("【%s】查询历史分时成交报错,%v", stCode, err))
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
		slog.Info(fmt.Sprintf("【%s】查询今日分时成交报错,%v", stCode, err))
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

// 获取本年度A股的所有节假日：https://www.tdx.com.cn/url/holiday/
func TdxHolidys() (days []int, err error) {
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("https://www.tdx.com.cn/url/holiday/")
	if err != nil {
		return
	}
	defer response.Body.Close()
	dataByte, err := io.ReadAll(response.Body)
	if nil != err {
		return days, err
	}
	utf8Data, _, _ := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), dataByte)
	respStr := string(utf8Data)

	// <textarea id="data" style="display:none;"></textarea>
	re := regexp.MustCompile(`<textarea\s+id=\"data\".*>([\s\S]*?)</textarea>`)
	matches := re.FindAllStringSubmatch(respStr, -1)
	if len(matches) < 1 || len(matches[0]) <= 1 {
		return days, errors.New("未匹配到内容")
	}
	lines := strings.Split(matches[0][1], "\r\n")
	for _, v := range lines {
		if len(v) > 10 && strings.Contains(v, "|中国|") {
			days = append(days, tdx.StrInt2Int(v[0:8]))
		}
	}
	return days, err
}
