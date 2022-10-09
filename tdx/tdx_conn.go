package tdx

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// 响应头的长度 16位0x10=16
var H_LEN int = 0x10

func NewTdxConn(connName, addr string) (tdx *TdxConn, err error) {
	newTdx := &TdxConn{
		ConnName:  connName,
		Addr:      addr,
		Connected: false,
	}
	conn, err := net.Dial("tcp", newTdx.Addr)
	if nil != err {
		return newTdx, err
	}
	newTdx.Connected = true
	newTdx.NetConn = conn
	//登录
	loginErr := newTdx.tdxConnLogin()
	if loginErr != nil {
		return newTdx, loginErr
	}
	//维持心跳
	go newTdx.heartBeat()

	return newTdx, nil
}

// 连接关闭
func (newTdx *TdxConn) TdxConnClose() {
	newTdx.Connected = false
	newTdx.NetConn.Close()
}

// 登录
func (newTdx *TdxConn) tdxConnLogin() error {
	lgVo1, errorLogin1 := newTdx.SendData(CmdBytesLogin1())
	if nil != errorLogin1 {
		newTdx.Connected = false
		return errorLogin1
	} else {
		log.Printf("第一次登录返回【%s】\n", BytesToHexStr(lgVo1.HedData))
	}
	lgVo2, errorLogin2 := newTdx.SendData(CmdBytesLogin2())
	if nil != errorLogin2 {
		newTdx.Connected = false
		return errorLogin2
	} else {
		log.Printf("第二次登录返回【%s】\n", BytesToHexStr(lgVo2.HedData))
	}
	lgVo3, errorLogin3 := newTdx.SendData(CmdBytesLogin3())
	if nil != errorLogin3 {
		newTdx.Connected = false
		return errorLogin3
	} else {
		log.Printf("第三次登录返回【%s】\n", BytesToHexStr(lgVo3.HedData))
	}
	return nil
}

// 心跳维持
func (newTdx *TdxConn) heartBeat() {
	for {
		//每15秒发送一次心跳检查
		time.Sleep(27 * time.Second)
		_, err := newTdx.SendData(CmdHeartbeat())
		if nil != err {
			log.Printf("【%s】心跳检查失败\n", newTdx.ConnName)
			break
		} else {
			log.Printf("【%s】心跳检查成功\n", newTdx.ConnName)
		}
		if !newTdx.Connected {
			break
		}
	}
}

// 请求数据
func (tc *TdxConn) SendData(data []byte) (vo TdxTcpPackVo, err error) {
	if !tc.Connected {
		return TdxTcpPackVo{}, errors.New("未建立连接")
	}
	tc.LockConn.Lock()
	defer tc.LockConn.Unlock()
	tc.NetConn.Write(data)
	//发送数据以后，马上开始接收
	return tc.reciveTcpPackVo()
}

// 发送报文后，接收数据
func (tc *TdxConn) reciveTcpPackVo() (vo TdxTcpPackVo, err error) {
	cacheBytes := []byte{}
	var resultVo TdxTcpPackVo
	var resultErr error
	for {
		buf := make([]byte, 1024)
		readLen, readError := tc.NetConn.Read(buf) // 从conn中读取
		if nil != readError {
			tc.Connected = false
			resultErr = readError
			break
		}
		cacheBytes = append(cacheBytes, buf[0:readLen]...)
		//如果读取的长度小于H长度，继续读
		if len(cacheBytes) < H_LEN {
			continue
		}
		headerVo := &TdxHeaderVo{}
		headerData := cacheBytes[0:H_LEN]
		BytesToVo(headerData, headerVo, true)
		if len(cacheBytes)-H_LEN < int(headerVo.Zipsize) {
			//包还不完整，所以还要继续读
			continue
		}
		//有完整的包了,先将cacheBytes里面的完整包取出来，然后把剩下的数据继续赋值给cacheBytes
		bodyData := cacheBytes[H_LEN : int(headerVo.Zipsize)+H_LEN]
		cacheBytes = cacheBytes[H_LEN+int(headerVo.Zipsize):]
		//判断是否要ZIP解压
		resultVo, resultErr = tc.bodyBytesPkgVo(headerVo, headerData, bodyData)
		//读取到了一个完整的包了，不需要再次读取，因为TDX通信协议一次只返回一个包
		if len(cacheBytes) > 0 {
			log.Printf("通达信返回的数据包，读取一个完整的包以后还有其他字节，其他字节长度为【%d】\n", len(cacheBytes))
		}
		break
	}
	return resultVo, resultErr
}

// 将读取到的数据和HeadVo打包成TcpPackVo
func (tc *TdxConn) bodyBytesPkgVo(headerVo *TdxHeaderVo, headerBytes, bodyData []byte) (vo TdxTcpPackVo, err error) {
	pakageData := TdxTcpPackVo{Hedvo: *headerVo, HedData: headerBytes, BodyData: bodyData}
	//开始解包
	if headerVo.Zipsize == headerVo.Unzipsize {
		//不需要解压
		return pakageData, nil
	}
	//需要ZIP解包
	nr := bytes.NewReader(bodyData)
	zipReader, err := zlib.NewReader(nr)
	if nil != err {
		return pakageData, err
	}
	var out bytes.Buffer
	io.Copy(&out, zipReader)
	zipReader.Close()
	bodyData = out.Bytes()
	if headerVo.Unzipsize != int16(len(bodyData)) {
		return pakageData, errors.New("解压后的包长度与报文头中的长度不相等")
	}
	//因为解压了数据， 所以重新赋值
	pakageData.BodyData = bodyData
	return pakageData, nil
}

type TdxConn struct {
	ConnName  string
	Addr      string
	Connected bool
	NetConn   net.Conn
	LockConn  sync.Mutex
}

type TdxHeaderVo struct {
	H1      uint16
	H2      uint16
	H3      uint16
	H4      uint16
	H5      uint16
	TdxFunc uint16
	//前三个short不要
	Zipsize   int16
	Unzipsize int16
}

type TdxTcpPackVo struct {
	Hedvo    TdxHeaderVo
	HedData  []byte
	BodyData []byte
}
