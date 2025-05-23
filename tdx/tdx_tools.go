package tdx

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/nie312122330/niexq-gotools/dateext"
	"github.com/shopspring/decimal"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func DataReadDate(data []byte, inPos *int, dataSize int) (y, mon, d int) {
	zipday := uint16(0)
	BytesToVo(data[*inPos:*inPos+dataSize], &zipday, true)
	year := (zipday >> 11) + 2004
	month := int((zipday % 2048) / 100)
	day := (zipday % 2048) % 100
	*inPos = *inPos + dataSize
	return int(year), month, int(day)
}

func DataReadTime(data []byte, inPos *int, dataSize int) (h, m int32) {
	time := int16(0)
	BytesToVo(data[*inPos:*inPos+dataSize], &time, true)
	*inPos = *inPos + dataSize
	return int32(time / 60), int32(time % 60)
}

func DataReadFloat(data []byte, inPos *int, dataSize int) (rst float32) {
	vol := float32(0)
	BytesToVo(data[*inPos:*inPos+dataSize], &vol, true)
	*inPos = *inPos + dataSize
	return vol
}

func DataReadint32(data []byte, inPos *int) (rst int32) {
	num := int32(0)
	BytesToVo(data[*inPos:*inPos+4], &num, true)
	*inPos = *inPos + 4
	return num
}

func DataReadint16(data []byte, inPos *int) (rst int16) {
	num := int16(0)
	BytesToVo(data[*inPos:*inPos+2], &num, true)
	*inPos = *inPos + 2
	return num
}

func DataReadint8(data []byte, inPos *int) (rst int8) {
	num := int8(0)
	BytesToVo(data[*inPos:*inPos+1], &num, true)
	*inPos = *inPos + 1
	return num
}

func DataReaduint8(data []byte, inPos *int) (rst uint8) {
	num := uint8(0)
	BytesToVo(data[*inPos:*inPos+1], &num, true)
	*inPos = *inPos + 1
	return num
}

// 这个很特殊-读取有符号和无符号数据（+-）
func DataReadSignNum(data []byte, pos *int) int {
	pos_byte := 6
	bdata := int(data[*pos])
	*pos = *pos + 1
	intdata := bdata & 0x3F

	//确定是否有符号
	sign := false
	if bdata&0x40 > 0 {
		sign = true
	} else {
		sign = false
	}

	if bdata&0x80 > 0 {
		for {
			bdata = int(data[*pos])
			*pos = *pos + 1

			intdata += (bdata & 0x7f) << pos_byte
			pos_byte += 7
			if bdata&0x80 > 0 {
				//
			} else {
				break
			}
		}
	}

	if sign {
		intdata = -intdata
	}
	//PYTDX中有判断符号的地方， 为什么要判断，没搞懂
	return int(intdata)
}

func HexStrToBytes(hexStr string) []byte {
	outData, _ := hex.DecodeString(hexStr)
	return outData
}

func BytesToHexStr(data []byte) string {
	inDataHex := hex.EncodeToString(data)
	return inDataHex
}

func NumToBytes[T byte | int8 | uint16 | int16 | int32 | uint32](num T, minOrder bool) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	if minOrder {
		binary.Write(bytebuf, binary.LittleEndian, num)
	} else {
		binary.Write(bytebuf, binary.BigEndian, num)
	}
	return bytebuf.Bytes()
}

func BytesToVo(byteData []byte, refVo interface{}, minOrder bool) {
	if minOrder {
		binary.Read(bytes.NewBuffer(byteData), binary.LittleEndian, refVo)
	} else {
		binary.Read(bytes.NewBuffer(byteData), binary.BigEndian, refVo)
	}
}

func BubbleSort(arr *[]int32) {
	temp := int32(0)
	for i := 0; i < len(*arr)-1; i++ {
		for j := 0; j < len(*arr)-1-i; j++ {
			if (*arr)[j] > (*arr)[j+1] {
				temp = (*arr)[j]
				(*arr)[j] = (*arr)[j+1]
				(*arr)[j+1] = temp
			}
		}
	}
}

func TimeStr2Time(timeStr string) time.Time {
	start, err := time.Parse(TIME_LAYOUT, timeStr)
	if nil != err {
		panic(err)
	}
	return start
}

func StrFloat2Float(str string) float64 {
	r, err := strconv.ParseFloat(str, 64)
	if nil != err {
		panic(err)
	}
	return r
}

func StrInt2Int(str string) int {
	r, err := strconv.ParseInt(str, 10, 32)
	if nil != err {
		panic(err)
		//1138  1143 1150 1154 1157 1159 1160 1161
	}
	return int(r)
}

func StrInt2Int64(str string) int64 {
	r, err := strconv.ParseInt(str, 10, 64)
	if nil != err {
		panic(err)
		//1138  1143 1150 1154 1157 1159 1160 1161
	}
	return r
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func MktByStCode(stCode string) int16 {
	if strings.HasPrefix(stCode, "00") || strings.HasPrefix(stCode, "30") {
		return 0
	} else if strings.HasPrefix(stCode, "60") || strings.HasPrefix(stCode, "68") {
		return 1
	}
	return -1
}

// 计算涨停价
func ZtPrice(preClosePrice int, ztfd float64) int {
	ztMoneyFloat := float64(preClosePrice) / float64(100) * (ztfd + 1)
	ztmoney := decimal.NewFromFloat(ztMoneyFloat)
	ztPrice, _ := ztmoney.Round(2).Float64()
	intStr := fmt.Sprintf("%1.0f", ztPrice*100)
	i, _ := strconv.Atoi(intStr)
	return i
}

// IF判断
func TdxIf[T any](cond bool, ok, no T) T {
	if cond {
		return ok
	} else {
		return no
	}
}

// 分时图计算当前时间
func TdxCalFsTime(year, month, day int, idx int) dateext.Date {
	// 09:30:00= 9*60+30=570
	// 13:00:00= 13*60=780  780-120=660 因为中间间隔了2个小时，所以要减去120分钟
	hour := TdxIf(idx < 120, (570+idx), (660+idx)) / 60
	minu := TdxIf(idx < 120, (570+idx), (660+idx)) % 60
	send := 0
	return dateext.WithDate(year, month, day, hour, minu, send)
}

// 分时图计算当前时间
func TdxCalFsTimeByDayInt(dayInt int, idx int) dateext.Date {
	dateStr := fmt.Sprintf("%d", dayInt)
	year := StrInt2Int(dateStr[0:4])
	month := StrInt2Int(dateStr[4:6])
	day := StrInt2Int(dateStr[6:8])
	return TdxCalFsTime(year, month, day, idx)
}

// 小数转整数
func Float2Int(v float64) int {
	ints := decimal.NewFromFloat(v).Round(0).String()
	return StrInt2Int(ints)
}

// 小数转整数
func Float2Int64(v float64) int64 {
	ints := decimal.NewFromFloat(v).Round(0).String()
	return StrInt2Int64(ints)
}
