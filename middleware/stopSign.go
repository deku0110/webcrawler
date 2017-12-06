package middleware

import (
	"sync"
	"fmt"
)

type StopSign interface {
	//发出停止信号
	Sign() bool
	//是否发出过停止信号
	Signed() bool
	//重置停止信号,相当于回收停止信号,并清楚所有的停止信号处理记录
	Reset()
	//处理停止信号
	//code应该代表停止信号处理方的代号,该带好出现在停止信号的处理记录中
	Deal(code string)
	//获得某个停止信号处理方的处理计数,该处理计数会从相应的停止信号处理记录中获得
	DealCount(code string) uint32
	//获取停止信号被处理的总计数
	DealTotal() uint32
	//获取摘要信息,其中应该包含所有的停止信号处理记录
	Summary() string
}

type myStopSign struct {
	signed       bool              //是否已经发出
	dealCountMap map[string]uint32 //处理计数字典
	rwmutex      sync.RWMutex      //读写锁
}

func NewStopSign() StopSign{
	ss := &myStopSign{
		dealCountMap:make(map[string]uint32),
	}
	return ss
}

func (ss *myStopSign) Sign() bool {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if ss.signed {
		return false
	}
	ss.signed = true
	return true
}

func (ss *myStopSign) Signed() bool {
	return ss.signed
}

func (ss *myStopSign) Reset() {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	ss.signed = false
	ss.dealCountMap = make(map[string]uint32)
}

func (ss *myStopSign) Deal(code string) {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if !ss.signed {
		return
	}
	if _,ok := ss.dealCountMap[code]; !ok{
		ss.dealCountMap[code] = 1
	}else {
		ss.dealCountMap[code] += 1
	}
}

func (ss *myStopSign) DealCount(code string) uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	return ss.dealCountMap[code]
}

func (ss *myStopSign) DealTotal() uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	var total uint32
	for _,v := range ss.dealCountMap {
			total += v
	}
	return total
}

func (ss *myStopSign) Summary() string {
	if ss.signed {
		return fmt.Sprintf("signed: true, dealCount: %v", ss.dealCountMap)
	} else {
		return "signed: false"
	}
}
