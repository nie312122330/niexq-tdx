package tdx

import (
	"errors"
	"fmt"
	"io"

	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/nie312122330/niexq-gotools/fileext"
)

// 股票列表-今日行情的基础数据
func (tc *TdxConn) QueryTodayStList(mkt, start uint16) []StListItemVo {
	enc := mahonia.NewDecoder("GBK")
	vo, _ := tc.SendData(CmdStList(mkt, start))
	dataCount := int16(0)
	BytesToVo(vo.BodyData[0:2], &dataCount, true)
	//每一个股票29个字节
	vos := []StListItemVo{}
	pos := 2
	for i := int16(0); i < dataCount; i++ {
		// fmt.Printf("%v\n", vo.BodyData[pos:pos+29])
		//<6sH8s4sBI4s
		stCode := enc.ConvertString(string(vo.BodyData[pos : pos+6]))
		pos += 6
		//单位[长度2]
		pos += 2
		// _ := DataReadint16(vo.BodyData, &pos)
		stName := enc.ConvertString(string(vo.BodyData[pos : pos+8]))
		pos += 8
		//未知字符[长度4]
		pos += 4
		// _42 := DataReadint8(vo.BodyData, &pos)
		//单位[长度1]
		pos += 1
		// dcimalPoint := DataReadint8(vo.BodyData, &pos)
		pre_price := DataReadFloat(vo.BodyData, &pos, 4)
		//最后4位，不知道是什么，+4过滤
		pos += 4
		// fmt.Printf("%v,%v,%0.2f\n", stCode, stName, pre_price)
		vos = append(vos, StListItemVo{
			StCode:   stCode,
			StName:   stName,
			PreClose: Float2Int(float64(pre_price * 100)),
		})
	}
	return vos
}

// 查询今日集合竞价
func (tc *TdxConn) QueryTodayJhjj(mkt int16, stCode string) (resuls *TdxRespBaseVo[TdxJhjjVo], err error) {
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
		h, m := DataReadTime(vo.BodyData, &pos, 2)                             //16-2  时间：hh:mm
		money := Float2Int(float64(DataReadFloat(vo.BodyData, &pos, 4) * 100)) //14-4  价格：分
		vol := DataReadint32(vo.BodyData, &pos)                                //10-4  匹配量
		unvol := DataReadint32(vo.BodyData, &pos)                              //6-4   未匹配量(有+-)
		unData := DataReadint8(vo.BodyData, &pos)                              //跳过1位,不晓得有什么用 2-1
		sec := DataReadint8(vo.BodyData, &pos)                                 //这一位是时间的秒 1-1  时间：秒

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

// 查询今日分时成交-分页
func (tc *TdxConn) QueryTodayPageFscj(mkt int16, stCode string, preClosePrice int, startPos, endPost int16) (resuls *TdxRespBaseVo[TdxFscjVo], err error) {
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
		if preClosePrice == -1 {
			preClosePrice = price_raw
		}
		last_price = last_price + price_raw

		datas = append(datas, TdxFscjVo{
			Hour:      int(h),
			Minus:     int(m),
			Second:    int(0),
			Price:     int(last_price),
			Vol:       int(vol),
			Num:       int(num),
			Buyorsell: int(buyorsell),
			PreClose:  preClosePrice,
		})
	}
	//赋值
	resultVo.Datas = datas
	return resultVo, nil
}

// 查询历史分时成交-分页
func (tc *TdxConn) QueryLsPageFscj(date int32, mkt int16, stCode string, startPos, endPost int16) (resuls *TdxRespBaseVo[TdxFscjVo], err error) {
	resultVo := &TdxRespBaseVo[TdxFscjVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_LSFSCJ,
	}
	vo, err := tc.SendData(CmdLsFscj(date, mkt, stCode, startPos, endPost))
	if nil != err {
		return resultVo, err
	}

	dataCount := int16(0)
	BytesToVo(vo.BodyData[0:2], &dataCount, true)
	pos := 2
	//上一日的收盘价
	preClosePrice := Float2Int(float64(DataReadFloat(vo.BodyData, &pos, 4) * 100))
	//上一次的价格
	last_price := 0

	datas := []TdxFscjVo{}
	for i := int16(0); i < dataCount; i++ {
		//解析时间
		h, m := DataReadTime(vo.BodyData, &pos, 2)
		//解析金额
		price_raw := DataReadSignNum(vo.BodyData, &pos)
		//解析量
		vol := DataReadSignNum(vo.BodyData, &pos)
		//解析笔数  - 历史分时成交 没有笔数
		// num := DataReadSignNum(vo.BodyData, &pos)
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
			Num:       -1,
			Buyorsell: int(buyorsell),
			PreClose:  preClosePrice,
		})
	}
	//赋值
	resultVo.Datas = datas
	return resultVo, nil
}

// 今日分时行情--需要传入昨日的收盘价
func (tc *TdxConn) QueryTodayFshq(mkt byte, stCode string, todayInt int, preClosePirce int) (resuls *TdxRespBaseVo[TdxFshqVo], preClosePrice int, err error) {
	resultVo := &TdxRespBaseVo[TdxFshqVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_TODAY_FSHQ,
	}
	vo, err := tc.SendData(CmdTodayFshq(int16(mkt), stCode))
	if nil != err {
		return resultVo, 0, err
	}
	if len(vo.BodyData) < 2 {
		return resultVo, 0, nil
	}
	pos := 0
	dataCount := int(DataReadint32(vo.BodyData, &pos))
	if dataCount <= 0 {
		return resultVo, 0, errors.New("没有返回数据")
	}

	datas := []TdxFshqVo{}

	//今日开盘价
	openPrice := DataReadSignNum(vo.BodyData, &pos)
	//今日均价
	openAvgPrice := DataReadSignNum(vo.BodyData, &pos)
	//开盘的量
	openVol := DataReadSignNum(vo.BodyData, &pos)

	//记录当前价格
	lastPrice := openPrice
	datas = append(datas, TdxFshqVo{
		DateTime: TdxJsonTime(TdxCalFsTimeByDayInt(todayInt, 0).Time),
		Price:    openPrice,
		AvgPrice: openAvgPrice / 100,
		Vol:      openVol,
		VolFlag:  TdxIf(openPrice > preClosePrice, 1, TdxIf(openPrice == preClosePrice, 0, -1)),
	})

	for i := 1; i < dataCount; i++ {
		changePrice := DataReadSignNum(vo.BodyData, &pos)
		changeAvgPrice := DataReadSignNum(vo.BodyData, &pos)
		curVol := DataReadSignNum(vo.BodyData, &pos)
		curPrice := openPrice + changePrice
		curAvgPrice := (openAvgPrice + changeAvgPrice) / 100
		vo := TdxFshqVo{
			DateTime: TdxJsonTime(TdxCalFsTimeByDayInt(todayInt, i).Time),
			Price:    curPrice,
			AvgPrice: curAvgPrice,
			Vol:      curVol,
			VolFlag:  TdxIf(curPrice > lastPrice, 1, TdxIf(curPrice == lastPrice, 0, -1)),
		}
		if i == 239 {
			//强制最后一根为白色
			vo.VolFlag = 0
		}
		datas = append(datas, vo)
		lastPrice = curPrice
	}
	resultVo.Datas = datas
	return resultVo, preClosePirce, nil
}

// 历史分时行情【均价的计算方式不太正确，有偏离，但是偏离不大，基本可用，有可能是0fb4这个命令是之前的功能，不再深入研究了】
func (tc *TdxConn) QueryLsFshq(date int32, mkt byte, stCode string) (resuls *TdxRespBaseVo[TdxFshqVo], preClosePrice int, err error) {
	resultVo := &TdxRespBaseVo[TdxFshqVo]{
		Market:  int(mkt),
		StCode:  stCode,
		TdxFunc: TDX_FUNC_LSFSHQ,
	}
	vo, err := tc.SendData(CmdFshq(date, mkt, stCode))
	if nil != err {
		return resultVo, 0, err
	}
	if len(vo.BodyData) < 2 {
		return resultVo, 0, nil
	}

	pos := 0
	dataCount := int(DataReadint16(vo.BodyData, &pos))
	if dataCount <= 0 {
		return resultVo, 0, errors.New("没有返回数据")
	}
	//读取上日的收盘价
	closePrice := Float2Int(float64(DataReadFloat(vo.BodyData, &pos, 4) * 100))

	datas := []TdxFshqVo{}
	//今日开盘价
	openPrice := DataReadSignNum(vo.BodyData, &pos)
	//今日均价
	//均价的计算方式不太正确，有偏离，但是偏离不大，基本可用，有可能是0fb4这个命令是之前的功能，不再深入研究了
	openAvgPrice := DataReadSignNum(vo.BodyData, &pos)
	//开盘的量
	openVol := DataReadSignNum(vo.BodyData, &pos)

	//记录当前价格
	lastPrice := openPrice
	lastAvgPrice := openAvgPrice
	datas = append(datas, TdxFshqVo{
		DateTime: TdxJsonTime(TdxCalFsTimeByDayInt(int(date), 0).Time),
		Price:    openPrice,
		AvgPrice: openPrice + (lastAvgPrice / 100),
		Vol:      openVol,
		VolFlag:  TdxIf(openPrice > preClosePrice, 1, TdxIf(openPrice == preClosePrice, 0, -1)),
	})
	for i := 1; i < dataCount; i++ {
		changePrice := DataReadSignNum(vo.BodyData, &pos)
		changeAvgPrice := DataReadSignNum(vo.BodyData, &pos)
		curVol := DataReadSignNum(vo.BodyData, &pos)

		curPrice := lastPrice + changePrice
		lastAvgPrice = lastAvgPrice + changeAvgPrice
		vo := TdxFshqVo{
			DateTime: TdxJsonTime(TdxCalFsTimeByDayInt(int(date), i).Time),
			Price:    curPrice,
			AvgPrice: openPrice + (lastAvgPrice / 100),
			Vol:      curVol,
			VolFlag:  TdxIf(curPrice > lastPrice, 1, TdxIf(curPrice == lastPrice, 0, -1)),
		}
		if i == 239 {
			//强制最后一根为白色
			vo.VolFlag = 0
		}
		datas = append(datas, vo)
		lastPrice = curPrice
	}

	//赋值
	resultVo.Datas = datas
	return resultVo, closePrice, nil
}

// 1分钟的K线
func (tc *TdxConn) QueryLsBarK1m(mkt int16, stCode string, start, count int16) (resuls *TdxRespBaseVo[TdxBarK1mVo], err error) {
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

		vol_f := DataReadFloat(vo.BodyData, &pos, 4)
		money_f := DataReadFloat(vo.BodyData, &pos, 4)
		//数字保存
		volInt := int64(Float2Int(float64(vol_f)))
		money := int64(Float2Int(float64(money_f * 100)))

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
	dataByte, err := io.ReadAll(response.Body)
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

// 读取GBK文件，并转为U8
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
