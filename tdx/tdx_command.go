package tdx

import (
	"time"
)

// public int LOGIN_ONE                            = 0x000d;//第一次登录
// public int LOGIN_TWO                            = 0x0fdb;//第二次登录
// public int HEART                                = 0x0004;//心跳维持
// public int STOCK_COUNT                          = 0x044e;//股票数目
// public int STOCK_LIST                           = 0x0450;//股票列表
// public int KMINUTE                              = 0x0537;//当天分时K线
// public int KMINUTE_OLD                          = 0x0fb4;//指定日期分时K线
// public int KLINE                                = 0x052d;//股票K线
// public int BIDD                                 = 0x056a;//当日的竞价
// public int QUOTE                                = 0x053e;//实时五笔报价
// public int QUOTE_SORT                           = 0x053e;//沪深排序
// public int TRANSACTION                          = 0x0fc5;//分笔成交明细
// public int TRANSACTION_OLD                      = 0x0fb5;//历史分笔成交明细
// public int FINANCE                              = 0x0010;//财务数据
// public int COMPANY                              = 0x02d0;//公司数据  F10
// public int EXDIVIDEND                           = 0x000f;//除权除息
// public int FILE_DIRECTORY                       = 0x02cf;//公司文件目录
// public int FILE_CONTENT                         = 0x02d0;//公司文件内容

const TDX_FUNC_LOGIN uint16 = 0x000d
const TDX_FUNC_LOGIN2 uint16 = 0x0fdb
const TDX_FUNC_HEART_BEAT uint16 = 0x0004
const TDX_FUNC_KBAR uint16 = 0x052d
const TDX_FUNC_LSFSHQ uint16 = 0x0fb4
const TDX_FUNC_TODAY_FSHQ uint16 = 0x0537 //0x051d
const TDX_FUNC_FSCJ uint16 = 0x0fc5
const TDX_FUNC_LSFSCJ uint16 = 0x0fb5
const TDX_FUNC_JHJY uint16 = 0x056a

func CmdBytesLogin1() []byte {
	// cmdHex1 := "0c0218930001030003000d0001"
	outData := HexStrToBytes("0c021893000103000300") //0d00
	outData = append(outData, NumToBytes(TDX_FUNC_LOGIN, true)...)
	outData = append(outData, NumToBytes(int8(0x01), true)...)
	return outData
}

func CmdBytesLogin2() []byte {
	// cmdHex1 := "0c0218940001030003000d0002"
	outData := HexStrToBytes("0c021894000103000300") //000d0002
	outData = append(outData, NumToBytes(TDX_FUNC_LOGIN, true)...)
	outData = append(outData, NumToBytes(int8(0x02), true)...)
	return outData
}

func CmdBytesLogin3() []byte {
	// cmdHex1 := "0c031899000120002000db0fd5d0c9ccd6a4a8af0000008fc22540130000d500c9ccbdf0d7ea00000002"
	outData := HexStrToBytes("0c031899000120002000db0fd5d0c9ccd6a4a8af0000008fc22540130000d500c9ccbdf0d7ea00000002") //000d0002
	// outData = append(outData, NumToBytes(TDX_FUNC_LOGIN, true)...)
	// outData = append(outData, NumToBytes(int8(0x02), true)...)
	return outData
}

func CmdHeartbeat() []byte {
	// cmdHex1 := "b1cb74000c9101280000"
	//还不知道这个是不是心跳维持包
	outData := HexStrToBytes("0c")
	sec := time.Now().Second()
	outData = append(outData, NumToBytes(int8(sec), true)...)
	outData = append(outData, HexStrToBytes("0128000202000200")...)
	outData = append(outData, NumToBytes(TDX_FUNC_HEART_BEAT, true)...)
	return outData
}

// 当天的集合竞价
func CmdJhjj(mkt int16, stCode string) []byte {
	outData := HexStrToBytes("0c02080301011e001e00") //c50f
	outData = append(outData, NumToBytes(TDX_FUNC_JHJY, true)...)
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, []byte(stCode)...)
	//后面这一截都不知道什么意思
	outData = append(outData, HexStrToBytes("00000000030000000000000000000000f4010000")...)
	return outData
}

// 当天的分时成交
func CmdFscj(mkt int16, stCode string, startPos, endPos int16) []byte {
	outData := HexStrToBytes("0c17080101010e000e00") //c50f
	outData = append(outData, NumToBytes(TDX_FUNC_FSCJ, true)...)
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, []byte(stCode)...)
	outData = append(outData, NumToBytes(startPos, true)...)
	outData = append(outData, NumToBytes(endPos, true)...)
	return outData
}

// 历史分时成交
func CmdLsFscj(date int32, mkt int16, stCode string, startPos, endPos int16) []byte {
	//0c 01 30 01 00 01 12 00 12 00 b5 0f
	outData := HexStrToBytes("0c013001000112001200") //c50f
	outData = append(outData, NumToBytes(TDX_FUNC_LSFSCJ, true)...)
	outData = append(outData, NumToBytes(date, true)...)
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, []byte(stCode)...)
	outData = append(outData, NumToBytes(startPos, true)...)
	outData = append(outData, NumToBytes(endPos, true)...)
	return outData
}

// 今日分时行情-可传入日期
func CmdTodayFshq(mkt int16, stCode string) []byte {
	outData := HexStrToBytes("0c1b080001010e000e00")
	outData = append(outData, NumToBytes(TDX_FUNC_TODAY_FSHQ, true)...)
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, []byte(stCode)...)
	outData = append(outData, NumToBytes(uint32(0), true)...)
	return outData
}

// 历史分时行情-可传入日期
func CmdFshq(date int32, mkt byte, stCode string) []byte {
	outData := HexStrToBytes("0c01300001010d000d00")
	outData = append(outData, NumToBytes(TDX_FUNC_LSFSHQ, true)...)
	outData = append(outData, NumToBytes(date, true)...)
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, []byte(stCode)...)
	return outData
}

// K线1分钟
// CmdK1mbar( 0, "000630", 0, 10)
func CmdBarK1m(mkt int16, stCode string, start, count int16) []byte {
	outData := NumToBytes(int16(0x10c), true)
	outData = append(outData, NumToBytes(int32(0x01016408), true)...)
	outData = append(outData, NumToBytes(int16(0x1c), true)...)
	outData = append(outData, NumToBytes(int16(0x1c), true)...)
	outData = append(outData, NumToBytes(TDX_FUNC_KBAR, true)...)
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, []byte(stCode)...)
	outData = append(outData, NumToBytes(int16(7), true)...) //1分钟K线
	outData = append(outData, NumToBytes(int16(1), true)...)
	outData = append(outData, NumToBytes(start, true)...)
	outData = append(outData, NumToBytes(count, true)...)
	outData = append(outData, NumToBytes(int32(0), true)...)
	outData = append(outData, NumToBytes(int32(0), true)...)
	outData = append(outData, NumToBytes(int16(0), true)...)
	return outData
}

// 获取股票列表
// CmdStList( 0, 0)
func CmdStList(mkt uint16, start uint16) []byte {
	outData := HexStrToBytes("0c0118640101060006005004") //c50f
	outData = append(outData, NumToBytes(mkt, true)...)
	outData = append(outData, NumToBytes(start, true)...)
	return outData
}
