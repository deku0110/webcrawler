package itempipeline

import (
	"webcrawler/base"
	"errors"
	"fmt"
	"sync/atomic"
)

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

type myItemPipeLine struct {
	itemProcesors    []ProcessItem
	failFast         bool
	send             uint64 //已发送条目的数量
	accepted         uint64 //已接受数量
	processed        uint64 //已处理条目数量
	processingNumber uint64 //处理中数量
}

func NewItemPipeline(itemProcessors []ProcessItem) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New(fmt.Sprintf("Invalid item processor list!")))
	}
	innerItemProcessors := make([]ProcessItem,0)
	for i,ip := range itemProcessors {
		if ip == nil {
			panic(errors.New(fmt.Sprintf("Invalid item processor[%d]!\n",i)))
		}
		innerItemProcessors = append(innerItemProcessors,ip)
	}
	items := &myItemPipeLine{itemProcesors: innerItemProcessors}
	return items
}

func (it *myItemPipeLine) Send(item base.Item) []error {
	atomic.AddUint64(&it.processingNumber,1)
	defer atomic.AddUint64(&it.processingNumber,^uint64(0))
	atomic.AddUint64(&it.send,1)
	errs := make([]error,0)
	if item == nil {
		errs = append(errs,errors.New("The item is invalid!"))
		return errs
	}
	atomic.AddUint64(&it.accepted,1)
	var currentItem base.Item = item
	for _,itemProcessor := range it.itemProcesors {
		processedItem,err := itemProcessor(currentItem)
		if err != nil {
			errs = append(errs,err)
			if it.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	atomic.AddUint64(&it.processed,1)
	return errs
}

func (it *myItemPipeLine) FailFast() bool {
	return it.failFast
}

func (it *myItemPipeLine) SetFailFast(failFast bool) {
	it.failFast = failFast
}

func (it *myItemPipeLine) Count() []uint64 {
	counts := make([]uint64, 3)
	counts[0] = atomic.LoadUint64(&it.send)
	counts[1] = atomic.LoadUint64(&it.accepted)
	counts[2] = atomic.LoadUint64(&it.processed)
	return counts
}

func (it *myItemPipeLine) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&it.processingNumber)
}

var summaryTemplate = "failFast: %v, processorNumber: %d," +
	" sent: %d, accepted: %d, processed: %d, processingNumber: %d"

func (it *myItemPipeLine) Summary() string {
	counts := it.Count()
	summary := fmt.Sprintf(summaryTemplate,
		it.failFast, len(it.itemProcesors),
		counts[0], counts[1], counts[2], it.ProcessingNumber())
	return summary
}
