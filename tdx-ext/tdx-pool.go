package tdxext

import (
	"errors"
	"fmt"
	"log"

	"sync"
	"sync/atomic"
	"time"

	"github.com/nie312122330/niexq-tdx/tdx"
)

// 连接池的大小
var maxConnNum = int32(10)
var connChan = make(chan *tdx.TdxConn, maxConnNum)
var curConnNum = int32(0)
var tdxHqAddr = "119.147.212.81:7709"

// 同步锁
var lock sync.Mutex

// 初始化连接池
func InitPool(maxNum int32, hqAddr string) {
	maxConnNum = maxNum
	tdxHqAddr = hqAddr
	connChan = make(chan *tdx.TdxConn, maxConnNum)
}

// 获取连接
func GetConn() (tdxConn *tdx.TdxConn, err error) {
	lock.Lock()
	defer lock.Unlock()
	return acGetConn()
}

func acGetConn() (tdxConn *tdx.TdxConn, err error) {
	fmt.Printf("当前空闲连接数为:【%d】,已创建连接数【%d】，最大连接数【%d】\n", len(connChan), curConnNum, maxConnNum)
	if len(connChan) > 0 || atomic.LoadInt32(&curConnNum) >= maxConnNum {
		//连接池有空闲连接或者创建的链接数已最大了，只能等待
		r := <-connChan
		if r.Connected {
			return r, nil
		} else {
			atomic.AddInt32(&curConnNum, -1)
			return acGetConn()
		}
	} else {
		return createConn()
	}
}

// 归还连接
func ReturnConn(tdxConn *tdx.TdxConn) {
	//归还连接的时候需要检查连接是否被关闭了,如果关闭了， 就不要归还回去
	if tdxConn.Connected {
		log.Printf("连接【%s】归还成功", tdxConn.ConnName)
		connChan <- tdxConn
	} else {
		atomic.AddInt32(&curConnNum, -1)
		log.Printf("连接【%s】已经被释放，不再放回", tdxConn.ConnName)
	}
}

// 创建连接
func createConn() (tdxConn *tdx.TdxConn, err1 error) {
	connName := fmt.Sprintf("N%d-%d", time.Now().UnixMilli(), time.Now().Nanosecond())
	tdxConn, err := tdx.NewTdxConn(connName, tdxHqAddr)
	if nil != err {
		log.Printf("连接【%s】通达信行情服务器【%s】出错:%v\n", connName, tdxHqAddr, err)
		return nil, errors.New("创建连接失败")
	} else {
		log.Printf("连接【%s】通达信行情服务器【%s】成功\n", connName, tdxHqAddr)
		atomic.AddInt32(&curConnNum, 1)
		return tdxConn, nil
	}
}
