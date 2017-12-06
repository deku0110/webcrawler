package scheduler

import (
	"webcrawler/base"
	"net/http"
	"webcrawler/analyzer"
	"webcrawler/itempipeline"
	"webcrawler/middleware"
	"webcrawler/downloader"
	"fmt"
	"github.com/kataras/golog"
	"errors"
	"sync/atomic"
)

//用来生成httpClient的方法
type GenHttpClient func() *http.Client

type Scheduler interface {
	Start(channelArgs base.ChannelArgs, poolBaseArgs base.PoolBaseArgs, crawDepth uint32,
		httpClientGenerator GenHttpClient, respParsers []analyzer.ParseResponse,
		item []itempipeline.ProcessItem, firstHttpReq *http.Request) (err error)

	Stop() bool

	Running() bool
	//错误通道,若为nil 表示通道不可用或者调度器处于停止状态
	ErrorChan() <-chan error
	//判断所有模块是否都处于空闲状态
	Idle() bool
	//摘要信息
	Summary(prefix string) SchedSummary
}

type myScheduler struct {
	channelArgs   base.ChannelArgs
	poolBaseArgs  base.PoolBaseArgs
	crawlDepth    uint32
	primaryDomain string //主域名
	chanman       middleware.ChannelManager
	stopSign      middleware.StopSign
	dlpool        downloader.PageDownloaderPool
	analyzerPool  analyzer.AnalyzerPool
	itemPipeline  itempipeline.ItemPipeline
	reqCahce      requestCache
	urlMap        map[string]bool
	running       uint32
}

func (sched *myScheduler) Start(channelArgs base.ChannelArgs, poolBaseArgs base.PoolBaseArgs, crawDepth uint32,
	httpClientGenerator GenHttpClient, respParsers []analyzer.ParseResponse,
	item []itempipeline.ProcessItem, firstHttpReq *http.Request) (err error) {
		defer func(){
			if p := recover(); p != nil {
				errMsg := fmt.Sprintf("Fatal Scheduler Error: %s\n", p)
				golog.Fatal(errMsg)
				err = errors.New(errMsg)
			}
		}()
		if atomic.LoadUint32(&sched.running,) == 1 {
			return errors.New("The scheduler has been started!\n")
		}
		if err := channelArgs.Check(); err != nil{
			return err
		}
		sched.channelArgs = channelArgs
		if err := poolBaseArgs.Check(); err != nil {
			return err
		}
		sched.poolBaseArgs = poolBaseArgs
		sched.crawlDepth = crawDepth

		sched.chanman = generateChannelManager(channelArgs)
		if httpClientGenerator == nil {
			return errors.New("The HTTP client generator list is invalid!")
		}
		dlpool, err := generatePageDownloaderPool(poolBaseArgs.PageDownloaderPoolSize(),httpClientGenerator)
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page downloader pool:%s\n",err)
		return errors.New(errMsg)
	}
	sched.dlpool = dlpool

	analyzerpool, err := generateAnalyzerPool(poolBaseArgs.AnalyzerPoolSize())
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page analyzer pool:%s\n",err)
		return errors.New(errMsg)
	}
	sched.analyzerPool = analyzerpool
	if item == nil {
		return errors.New("The item processor list is invalid")
	}
	for i,ip := range item {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
		}
	}
	sched.itemPipeline = generateItemPipeline(item)

	if sched.stopSign == nil {
		sched.stopSign = middleware.NewStopSign()
	}else{
		sched.stopSign.Reset()
	}
	sched.reqCahce = newRequestCache()
	sched.urlMap = make(map[string]bool)
	//TODO 未完待续

}

func (sched *myScheduler) Stop() bool {
	panic("implement me")
}

func (sched *myScheduler) Running() bool {
	panic("implement me")
}

func (sched *myScheduler) ErrorChan() <-chan error {
	panic("implement me")
}

func (sched *myScheduler) Idle() bool {
	panic("implement me")
}

func (sched *myScheduler) Summary(prefix string) SchedSummary {
	panic("implement me")
}

