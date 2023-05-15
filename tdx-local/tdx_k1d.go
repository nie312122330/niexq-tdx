package tdxlocal

import (
	"bytes"
	"encoding/binary"
	"os"
	"sort"
)

// 解析文件
func ParseStockLc1dFile(filePath string) []TdxLc1dVo {
	//确定文件名称
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
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
	return vos
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
