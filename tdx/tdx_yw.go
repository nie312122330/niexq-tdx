package tdx

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/nie312122330/niexq-gotools/fileext"
)

// 查询集合竞价
func (tc *TdxConn) QueryJhjj(mkt int16, stCode string) (resuls *TdxRespBaseVo[TdxJhjjVo], err error) {
	resultVo := &TdxRespBaseVo[TdxJhjjVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_JHJY,
	}
	vo, err := tc.SendData(CmdJhjj(mkt, stCode))
	if nil != err {
		return resultVo, err
	}
	if len(vo.BodyData) < 2 {
		return resultVo, nil
	}
	dataCount := int16(0)
	BytesToVo(vo.BodyData[0:2], &dataCount, true)
	pos := 2

	datas := []TdxJhjjVo{}
	for i := int16(0); i < dataCount; i++ {
		//解析时间
		h, m := DataReadTime(vo.BodyData, &pos, 2)              //16-2  时间：hh:mm
		money := int(DataReadFloat(vo.BodyData, &pos, 4) * 100) //14-4  价格：分
		vol := DataReadint32(vo.BodyData, &pos)                 //10-4  匹配量
		unvol := DataReadint32(vo.BodyData, &pos)               //6-4   未匹配量(有+-)
		unData := DataReadint8(vo.BodyData, &pos)               //跳过1位,不晓得有什么用 2-1
		sec := DataReadint8(vo.BodyData, &pos)                  //这一位是时间的秒 1-1  时间：秒

		unFlag := 0
		if unvol > 0 {
			unFlag = 1
		} else if unvol < 0 {
			unFlag = -1
		}

		datas = append(datas, TdxJhjjVo{
			Hour:   int(h),
			Minus:  int(m),
			Second: int(sec),
			Price:  int(money),
			Vol:    int(vol),
			UnVol:  int(math.Abs(float64(unvol))),
			UnFlag: unFlag,
			UnData: int(unData),
		})
	}
	//赋值
	resultVo.Datas = datas
	return resultVo, nil
}

// 查询分时成交---暂不使用，业务中没有具体的参考
func (tc *TdxConn) QueryFscj(mkt int16, stCode string, startPos, endPost int16) (resuls *TdxRespBaseVo[TdxFscjVo], err error) {
	resultVo := &TdxRespBaseVo[TdxFscjVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_FSCJ,
	}
	vo, err := tc.SendData(CmdFscj(mkt, stCode, startPos, endPost))
	if nil != err {
		return resultVo, err
	}

	dataCount := int16(0)
	BytesToVo(vo.BodyData[0:2], &dataCount, true)
	pos := 2
	last_price := 0

	datas := []TdxFscjVo{}
	for i := int16(0); i < dataCount; i++ {
		//解析时间
		h, m := DataReadTime(vo.BodyData, &pos, 2)
		//解析金额
		price_raw := DataReadSignNum(vo.BodyData, &pos)
		//解析量
		vol := DataReadSignNum(vo.BodyData, &pos)
		//解析笔数
		num := DataReadSignNum(vo.BodyData, &pos)
		//解析买卖，2-竞价，1-卖，0买
		buyorsell := DataReadSignNum(vo.BodyData, &pos)
		//移动位置
		DataReadSignNum(vo.BodyData, &pos)

		last_price = last_price + price_raw

		datas = append(datas, TdxFscjVo{
			Hour:      int(h),
			Minus:     int(m),
			Second:    int(0),
			Price:     int(last_price),
			Vol:       int(vol),
			Num:       int(num),
			Buyorsell: int(buyorsell),
		})
	}
	//赋值
	resultVo.Datas = datas
	return resultVo, nil
}

// 分时行情
func (tc *TdxConn) QueryFshq(date int32, mkt byte, stCode string) (resuls *TdxRespBaseVo[TdxFshqVo], preClosePrice int, err error) {
	resultVo := &TdxRespBaseVo[TdxFshqVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_FSHQ,
	}
	vo, err := tc.SendData(CmdFshq(date, mkt, stCode))
	if nil != err {
		return resultVo, 0, err
	}
	if len(vo.BodyData) < 2 {
		return resultVo, 0, nil
	}
	dataCount := int16(0)
	BytesToVo(vo.BodyData[0:2], &dataCount, true)
	pos := 2
	if dataCount <= 0 {
		return resultVo, 0, errors.New("没有返回数据")
	}
	dateStr := fmt.Sprintf("%d", date)
	dateStr = dateStr[0:4] + "-" + dateStr[4:6] + "-" + dateStr[6:8]

	datas := []TdxFshqVo{}
	//上一日的收盘价
	closePrice := FloatXNumToInt(float64(DataReadFloat(vo.BodyData, &pos, 4)), 100)
	//读取分时行情的方法

	readFshqFunc := func(pData []byte, pPos *int, curPrice int, first bool) (pirce, volFlag, vol, unKonwData int) {
		volFlag = 0
		price_raw := DataReadSignNum(pData, &pos)
		unKonwData = DataReadSignNum(pData, &pos)
		vol = DataReadSignNum(vo.BodyData, &pos)
		last_price := 0
		if first {
			//此时是收盘价
			if price_raw > curPrice {
				volFlag = 1
			} else if price_raw < curPrice {
				volFlag = -1
			}
			last_price = price_raw
		} else {
			if price_raw > 0 {
				volFlag = 1
			} else if price_raw < 0 {
				volFlag = -1
			}
			last_price = curPrice + price_raw
		}
		return last_price, volFlag, vol, unKonwData
	}
	//第一次读取
	curPrice, volFlag, vol, unKonwData := readFshqFunc(vo.BodyData, &pos, closePrice, true)
	curTime, _ := time.Parse(TIME_LAYOUT, fmt.Sprintf("%s 09:30:00", dateStr))
	datas = append(datas, TdxFshqVo{DateTime: TdxJsonTime(curTime), Price: curPrice, UnKonwData: unKonwData, Vol: vol, VolFlag: volFlag})
	//读取剩余的数据,因为第一条已经读取了，所以i=1

	for i := int16(1); i < dataCount; i++ {
		curPrice, volFlag, vol, unKonwData = readFshqFunc(vo.BodyData, &pos, curPrice, false)
		h := 0
		m := 0
		if i < 30 {
			h = 9
			m = int(i) + 30
		} else if i < 90 {
			h = 10
			m = int(i - 30)
		} else if i < 120 {
			h = 11
			m = int(i - 90)
		} else if i < 180 {
			h = 13
			m = int(i - 120)
		} else {
			h = 14
			m = int(i - 180)
		}
		curTimestr := fmt.Sprintf("%s %02d:%02d:00", dateStr, h, m)
		curTime, _ := time.Parse(TIME_LAYOUT, curTimestr)
		datas = append(datas, TdxFshqVo{DateTime: TdxJsonTime(curTime), Price: curPrice, UnKonwData: unKonwData, Vol: vol, VolFlag: volFlag})
	}
	//赋值
	resultVo.Datas = datas
	return resultVo, closePrice, nil
}

// 查询指定日期内的收盘价及分钟最大量
func (tc *TdxConn) QueryDatesMaxVolAndClosePrice(dates []int32, mkt byte, stCode string) (closePrice, maxVol int) {
	//排序
	BubbleSort(&dates)
	datas := []TdxFshqVo{}
	for _, v := range dates {
		res, _, _ := tc.QueryFshq(v, mkt, stCode)
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

// 1分钟的K线
func (tc *TdxConn) QueryBarK1m(mkt int16, stCode string, start, count int16) (resuls *TdxRespBaseVo[TdxBarK1mVo], err error) {
	resultVo := &TdxRespBaseVo[TdxBarK1mVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_KBAR,
	}
	vo, err := tc.SendData(CmdBarK1m(mkt, stCode, start, count))
	if nil != err {
		return resultVo, err
	}
	//数量
	dataCount := int16(0)
	BytesToVo(vo.BodyData[0:2], &dataCount, true)
	pos := 2
	datas := []TdxBarK1mVo{}
	pre_diff_base := 0

	for i := int16(0); i < dataCount; i++ {
		y, mon, d := DataReadDate(vo.BodyData, &pos, 2)
		h, m := DataReadTime(vo.BodyData, &pos, 2)

		price_open_diff := DataReadSignNum(vo.BodyData, &pos)
		price_close_diff := DataReadSignNum(vo.BodyData, &pos)
		price_high_diff := DataReadSignNum(vo.BodyData, &pos)
		price_low_diff := DataReadSignNum(vo.BodyData, &pos)

		vol := DataReadFloat(vo.BodyData, &pos, 4)
		dbVol := DataReadFloat(vo.BodyData, &pos, 4)
		//数字保存
		volInt := int64(vol)
		money := int64(dbVol * 100)

		//保存数据
		open := price_open_diff + pre_diff_base
		close := open + price_close_diff
		high := open + price_high_diff
		low := open + price_low_diff

		//保存基础对比
		pre_diff_base = close

		//解析为对象
		curTime, _ := time.Parse(TIME_LAYOUT, fmt.Sprintf("%04d-%02d-%02d %02d:%02d:00", y, mon, d, h, m))
		datas = append(datas, TdxBarK1mVo{DateTime: TdxJsonTime(curTime), Open: open / 10, Close: close / 10, High: high / 10, Low: low / 10, Vol: volInt, Money: money})
	}
	resultVo.Datas = datas
	return resultVo, nil
}

// 从ECB文件中读取股票代码
func ReadStCodeFromEcbFile(ecbFilePath string) []string {
	content, erro := fileext.ReadFileContent(ecbFilePath)
	if nil != erro {
		panic(erro)
	}
	results := strings.Split(content, "\r\n")
	stocksList := []string{}
	for _, v := range results {
		if len(v) > 0 {
			stocksList = append(stocksList, v)
		}
	}
	return stocksList
}

// 利用QQ股票获取股票名称
func QueryStName(mkt byte, stCode string, retryTimes int) (name string, err error) {
	mktStr := "sz"
	if mkt == 1 {
		mktStr = "sh"
	}
	reqUrl := fmt.Sprintf("https://qt.gtimg.cn/q=s_%s%s", mktStr, stCode)
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(reqUrl)
	if err != nil {
		if retryTimes > 0 {
			retryTimes--
			return QueryStName(mkt, stCode, retryTimes)
		} else {
			return "", err
		}
	}
	defer response.Body.Close()
	dataByte, err := ioutil.ReadAll(response.Body)
	if nil != err {
		return "", err
	}
	utf8Byte, err := GbkToUtf8(dataByte)
	if nil != err {
		return "", err
	}
	str := string(utf8Byte)
	strArry := strings.Split(str, "~")
	return strArry[1], nil
}

// 从通达信导出的Txt文件中读取股票数据【代码	名称 涨幅	现价 涨跌】
func ReadTdxExportTxtFile(txtFilePath string) []TdxTxtStVo {
	contentStr, erro := ReadGbKFile(txtFilePath)
	if nil != erro {
		panic(erro)
	}
	results := []TdxTxtStVo{}
	lines := strings.Split(contentStr, "\r\n")
	for _, v := range lines {
		if len(v) <= 0 {
			continue
		}
		if strings.HasPrefix(v, "代码") || strings.HasPrefix(v, "数据来源") {
			continue
		}
		colArray := strings.Split(v, "\t")
		stCode := colArray[0]
		mkt := int16(0)
		if strings.HasPrefix(stCode, "0") || strings.HasPrefix(stCode, "3") {
			mkt = int16(0)
		} else if strings.HasPrefix(stCode, "6") {
			mkt = int16(1)
		} else {
			mkt = int16(4)
		}
		stName := colArray[1]
		zf, _ := strconv.ParseFloat(colArray[2], 64)
		close, _ := strconv.ParseFloat(colArray[3], 64)
		zd, _ := strconv.ParseFloat(colArray[4], 64)

		results = append(results, TdxTxtStVo{
			StCode: stCode,
			StMkt:  mkt,
			StName: stName,
			StZf:   int(zf * 100),
			StXJ:   int(close * 100),
			StZd:   int(zd * 100),
		})
	}
	return results
}

func ReadGbKFile(filePath string) (str string, err error) {
	readDatas, erro := fileext.ReadFileByte(filePath)
	if nil != erro {
		return "", erro
	}
	if len(readDatas) <= 0 {
		return "", nil
	}
	enc := mahonia.NewDecoder("GBK")
	result := enc.ConvertString(string(readDatas))
	return result, nil
}
