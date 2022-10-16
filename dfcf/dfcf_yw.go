package dfcf

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	niexqext "github.com/nie312122330/niexq-gotools"
	"github.com/nie312122330/niexq-gotools/jsonext"
	"github.com/nie312122330/niexq-tdx/tdx"
)

// 获取时间范围内,单分钟最大的量和时间范围结束的收盘价
// PS:数据[MAXVOL]与通达信有出入-不予采信
func DcGetDateRangeMaxVolClosePrice(stcode string, start, end time.Time) (maxVol int, closeMoney int) {
	resultVos, err := DcQueryDay5Fshq(5, stcode)
	if nil != err {
		log.Printf("【%s】DFCF获取5日分时图错误:%s\n", stcode, err)
		return -1, -1
	}
	if len(resultVos) <= 0 {
		return -1, -1
	}
	cp := 0
	for _, v := range resultVos {
		if time.Time(v.DateTime).After(start) && time.Time(v.DateTime).Before(end) {
			if v.Vol > maxVol {
				maxVol = v.Vol
			}
			cp = v.Price
		}
	}
	return maxVol, cp
}

// 查询最近5天的分时行情
// PS:数据[VolFlag]与通达信有出入-不予采信
func DcQueryDay5Fshq(retryTimes int, stcode string) (results []DcFshqVo, resultErr error) {
	resultVos := []DcFshqVo{}
	baseUrl := "http://push2his.eastmoney.com/api/qt/stock/trends2/get?fields1=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f17&fields2=f51,f52,f53,f54,f55,f56,f57,f58"
	baseUrl += "&ndays=5&iscr=0&secid=%s.%s"
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(fmt.Sprintf(baseUrl, stcode[0:1], stcode[1:]))
	if err != nil {
		if retryTimes > 0 {
			retryTimes--
			return DcQueryDay5Fshq(retryTimes, stcode)
		} else {
			log.Printf("第【%d】次,获取【%s】分时图数据错误【%v】\n", retryTimes, stcode, err.Error())
			return resultVos, err
		}
	}
	defer response.Body.Close()
	dataByte, err := ioutil.ReadAll(response.Body)
	if nil != err {
		log.Printf("获取【%s】分时图数据错误,%v\n", stcode, err)
		return resultVos, err
	}
	dataStr := string(dataByte)

	m := DcK1mResultVo{}
	jsonext.ToObj(&dataStr, &m)
	dataLen := len(m.Data.Trends)
	if dataLen <= 0 {
		log.Printf("获取【%s】分时图无数据\n", stcode)
		return resultVos, err
	}
	//解析数据
	for i := 0; i < dataLen; i++ {
		//时间， 开盘，收盘，最高，最低，量，金额，均价
		// "2022-06-24 14:20,3.27,3.26,3.27,3.26,4229,1379063.00,3.257"
		strArr := strings.Split(m.Data.Trends[i], ",")
		barTime, _ := time.Parse("2006-01-02 15:04:05", strArr[0]+":00")
		//为了与通达信保持一致，认为去掉9:30分的那条数据，并将后面的数据都-1分钟
		if barTime.Hour() == 9 && barTime.Minute() == 30 {
			continue
		} else {
			barTime = barTime.Add(-time.Minute)
		}
		kpf := tdx.FloatXNumToInt(tdx.StrFloat2Float(strArr[1]), 100)
		spf := tdx.FloatXNumToInt(tdx.StrFloat2Float(strArr[2]), 100)
		voFlag := 0
		if spf > kpf {
			voFlag = 1
		} else if spf < kpf {
			voFlag = -1
		}
		vol := tdx.StrInt2Int(strArr[5])
		vo := DcFshqVo{DateTime: DcJsonTime(barTime), Price: spf, Vol: vol, VolFlag: voFlag, UnKonwData: 0}
		resultVos = append(resultVos, vo)
	}
	return resultVos, nil
}

// 东方财富，分时图实时监控
func DcK1mMonitor(idx int, stCode string, dataCh chan DcK1mNotifyVo) {
	k1mDataRegExp := regexp.MustCompile(`data: (\{\".*\})`)

	baseUrl := "http://%d.push2.eastmoney.com/api/qt/stock/trends2/sse"
	baseUrl += "?fields1=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f17&fields2=f51,f52,f53,f54,f55,f56,f57,f58"
	baseUrl += "&ut=%s&secid=%s.%s&wbp2u=|0|0|0|web&ndays=1&iscr=0"
	uuid := niexqext.UUID()

	response, err := http.Get(fmt.Sprintf(baseUrl, idx%10, uuid, stCode[0:1], stCode[1:]))
	if err != nil {
		log.Println("get error")
		return
	}
	resultStr := ""
	for {
		data := make([]byte, 5*1024*1024)
		n, err := response.Body.Read(data)
		if nil != err {
			log.Printf("读取数据异常，关闭【%s】的数据获取\n", stCode)
			dataCh <- DcK1mNotifyVo{
				StockCode: stCode,
				Datas:     nil,
			}
			break
		}
		str := string(data[:n])
		resultStr += str
		resultArray := k1mDataRegExp.FindStringSubmatch(resultStr)
		if len(resultArray) > 1 {
			for i := 1; i < len(resultArray); i++ {
				m := DcK1mRespVo{}
				jsonext.ToObj(&resultArray[i], &m)
				vo := DcK1mNotifyVo{
					StockCode: stCode,
					Datas:     m.Data.Trends,
				}
				dataCh <- vo
			}
			resultStr = ""
		}
	}
}
