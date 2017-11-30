package middleware

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
	//获得某个停止信号处理房的处理计数,该处理计数会从相应的停止信号处理记录中获得
	DealCount(code string) uint32
	//获取停止信号被处理的总计数
	DealTotal() uint32
	//获取摘要信息,其中应该包含所有的停止信号处理记录
	Summary() string
}
