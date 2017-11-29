package itempipeline

import "webcrawler/base"

type ItemPipeline interface {
	//发送条目
	Send(item base.Item) []error
	//是否快速失败
	FailFast() bool
	//设置快速失败
	SetFailFast(failFast bool)
	//获得已发送,已接受和已处理的条目的计数值
	//更确切地说,作为结果值得切片总忽悠3个元素值,分别代表已发送,已接受和已处理的计数
	Count() []uint64
	//正在被处理的条目的数量
	ProcessingNumber() uint64
	//摘要
	Summary() string
}
