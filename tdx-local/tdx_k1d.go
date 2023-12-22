package tdxlocal

import (
	"bytes"
	"encoding/binary"
	"os"
	"sort"

	"github.com/nie312122330/niexq-tdx/tdx"
)

// 计算涨停数量
func CountDayK10cm(baseDir, code string) (count int, err error) {
	path := Code2FilePath(baseDir, code, "lday")
	vos, err := ParseStockLc1dFile(path)
	if nil != err {
		return count, err
	}
	if len(vos) > 1 {
		return 0, nil
	}
	SortDayVosByDate(vos, false)
	for i := 0; i < len(vos); i++ {
		if int(vos[i].Close) == tdx.ZtPrice(int(vos[i+1].Close), 0.1) {
			count++
		} else {
			break
		}
	}
	return count, err
}

// 解析文件
func ParseStockLc1dFile(filePath string) (res []TdxLc1dVo, err error) {
	//确定文件名称
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	dataLen := len(data)
	vos := []TdxLc1dVo{}
	for i := 0; i < dataLen; i = i + 32 {
		lc1mData := data[i : i+32]
		buf := bytes.NewBuffer(lc1mData)
		data := &TdxLc1dVo{}
		binary.Read(buf, binary.LittleEndian, data)
		vos = append(vos, *data)
	}
	return vos, nil
}

func SortDayVosByDate(vos []TdxLc1dVo, asc bool) {
	sort.Slice(vos, func(i, j int) bool {
		if !asc {
			return vos[i].Date > vos[j].Date
		} else {
			return vos[i].Date < vos[j].Date
		}
	})
}

type TdxLc1dVo struct {
	Date                   int32
	Open, High, Low, Close int32
	Amount                 float32
	Vol                    int32
	UnData                 int32
}
