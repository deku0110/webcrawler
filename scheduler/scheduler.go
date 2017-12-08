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
	"time"
	"strings"
)

// 组件的统一代号。
const (
	DOWNLOADER_CODE   = "downloader"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "item_pipeline"
	SCHEDULER_CODE    = "scheduler"
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

//创建调度器
func NewScheduler() Scheduler {
	return &myScheduler{}
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
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Scheduler Error: %s\n", p)
			golog.Fatal(errMsg)
			err = errors.New(errMsg)
		}
	}()
	if atomic.LoadUint32(&sched.running, ) == 1 {
		return errors.New("The scheduler has been started!\n")
	}
	if err := channelArgs.Check(); err != nil {
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
	dlpool, err := generatePageDownloaderPool(poolBaseArgs.PageDownloaderPoolSize(), httpClientGenerator)
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page downloader pool:%s\n", err)
		return errors.New(errMsg)
	}
	sched.dlpool = dlpool

	analyzerpool, err := generateAnalyzerPool(poolBaseArgs.AnalyzerPoolSize())
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page analyzer pool:%s\n", err)
		return errors.New(errMsg)
	}
	sched.analyzerPool = analyzerpool
	if item == nil {
		return errors.New("The item processor list is invalid")
	}
	for i, ip := range item {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
		}
	}
	sched.itemPipeline = generateItemPipeline(item)

	if sched.stopSign == nil {
		sched.stopSign = middleware.NewStopSign()
	} else {
		sched.stopSign.Reset()
	}
	sched.reqCahce = newRequestCache()
	sched.urlMap = make(map[string]bool)

	sched.startDownloading()
	sched.activateAnalyzers(respParsers)
	sched.openItemPipeline()
	sched.schedule(10 * time.Millisecond)

	if firstHttpReq == nil {
		return errors.New("The first HTTP request is invalid!")
	}
	pd,err := getPrimaryDomain(firstHttpReq.Host)
	if err != nil {
		return err
	}
	sched.primaryDomain = pd

	firstReq := base.NewRequest(firstHttpReq,0)
	sched.reqCahce.put(firstReq)

	return nil
}

func (sched *myScheduler) Stop() bool {
	if atomic.LoadUint32(&sched.running) != 1 {
		return false
	}
	sched.stopSign.Sign()
	sched.chanman.Close()
	sched.reqCahce.close()
	atomic.StoreUint32(&sched.running, 2)
	return true
}

func (sched *myScheduler) Running() bool {
	return atomic.LoadUint32(&sched.running) == 1
}

func (sched *myScheduler) ErrorChan() <-chan error {
	if sched.chanman.Status() != middleware.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return sched.getErrorChan()
}

func (sched *myScheduler) Idle() bool {
	idleDlPool := sched.dlpool.Used() == 0
	idleAnalyzerPool := sched.analyzerPool.Used() == 0
	idleItemPipeline := sched.itemPipeline.ProcessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleItemPipeline {
		return true
	}
	return false
}

func (sched *myScheduler) Summary(prefix string) SchedSummary {
	return NewSchedSummary(sched, prefix)
}


func (sched *myScheduler) startDownloading() {
	go func() {
		for {
			req, ok := <-sched.getReqChan()
			if !ok {
				break
			}
			go sched.download(req)
		}
	}()
}

func (sched *myScheduler) download(req base.Request) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Download Error %s\n", p)
			golog.Fatal(errMsg)
		}
	}()
	download, err := sched.dlpool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Downloader pool error: %s", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.dlpool.Return(download)
		if err != nil {
			errMsg := fmt.Sprintf("Downloader pool error: %s", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()
	code := generateCode(DOWNLOADER_CODE, download.Id())
	respp, err := download.Download(req)
	if respp != nil {
		sched.sendResp(*respp, code)
	}
	if err != nil {
		sched.sendError(err, code)
	}
}
func (sched *myScheduler) sendResp(resp base.Response, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getRespChan() <- resp
	return true
}
// 发送条目。
func (sched *myScheduler) sendItem(item base.Item, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getItemChan() <- item
	return true
}
// 发送错误。
func (sched *myScheduler) sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType base.ErrorType
	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = base.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = base.ANALYZER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = base.ITEM_PROCESSOR_ERROR
	}
	cError := base.NewCrawlerError(errorType, err.Error())
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	go func() {
		sched.getErrorChan() <- cError
	}()
	return true
}

// 把请求存放到请求缓存。
func (sched *myScheduler) saveReqToCache(req base.Request, code string) bool {
	httpReq := req.HttpReq()
	if httpReq == nil {
		golog.Warn("Ignore the request! It's HTTP request is invalid!\n")
		return false
	}
	reqUrl := httpReq.URL
	if reqUrl == nil {
		golog.Warn("Ignore the request! It's url is is invalid!\n")
		return false
	}
	if strings.ToLower(reqUrl.Scheme) != "http" {
		golog.Warnf("Ignore the request! It's url scheme '%s', but should be 'http'!\n", reqUrl.Scheme)
		return false
	}
	if _, ok := sched.urlMap[reqUrl.String()]; ok {
		golog.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", reqUrl)
		return false
	}
	if pd, _ := getPrimaryDomain(httpReq.Host); pd != sched.primaryDomain {
		golog.Warnf("Ignore the request! It's host '%s' not in primary domain '%s'. (requestUrl=%s)\n",
			httpReq.Host, sched.primaryDomain, reqUrl)
		return false
	}
	if req.Depth() > sched.crawlDepth {
		golog.Warnf("Ignore the request! It's depth %d greater than %d. (requestUrl=%s)\n",
			req.Depth(), sched.crawlDepth, reqUrl)
		return false
	}
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.reqCahce.put(&req)
	sched.urlMap[reqUrl.String()] = true
	return true
}

//激活分析器
func(sched *myScheduler) activateAnalyzers(respParsers []analyzer.ParseResponse) {
	go func() {
		for {
			resp,ok := <-sched.getRespChan()
			if !ok {
				break
			}
			go sched.analyze(respParsers,resp)
		}
	}()
}

func(sched *myScheduler) analyze(respParsers []analyzer.ParseResponse,resp base.Response) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Analysis Error: %s\n",p)
			golog.Fatal(errMsg)
		}
	}()
	ana, err := sched.analyzerPool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.analyzerPool.Return(ana)
		if err != nil {
			errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()
	code := generateCode(ANALYZER_CODE,ana.Id())
	dataList,errs := ana.Analyzer(respParsers, resp)
	if dataList != nil {
		for _,data := range dataList {
			if data != nil {
				continue
			}
			switch d:= data.(type) {
			case *base.Request :
				sched.saveReqToCache(*d,code)
			case *base.Item:
				sched.sendItem(*d,code)
			default:
				errMsg := fmt.Sprintf("Unsupported data type '%T'! (value=%v)\n", d, d)
				sched.sendError(errors.New(errMsg), code)

			}
		}
	}
	if errs != nil {
		for _,err := range errs {
			sched.sendError(err,code)
		}
	}
}

func(sched *myScheduler) openItemPipeline(){
	go func() {
		sched.itemPipeline.SetFailFast(true)
		code := ITEMPIPELINE_CODE
		for item := range sched.getItemChan() {
			go func(item base.Item) {
				defer func() {
					if p := recover(); p != nil{
						errMsg := fmt.Sprintf("Fatal Item Processing Error: %s\n", p)
						golog.Fatal(errMsg)
					}
				}()
				errs := sched.itemPipeline.Send(item)
				if errs != nil {
					for _,err := range errs {
						sched.sendError(err,code)
					}
				}
			}(item)
		}
	}()
}

//调度,适当的搬运请求缓存中的请求到请求通道
func(sched *myScheduler) schedule(interval time.Duration) {
	go func() {
		for {
			if sched.stopSign.Signed() {
				sched.stopSign.Deal(SCHEDULER_CODE)
				return
			}
			remainder := cap(sched.getReqChan()) - len(sched.getReqChan())
			var temp *base.Request
			for remainder >0 {
				temp = sched.reqCahce.get()
				if temp == nil {
					break
				}
				if sched.stopSign.Signed() {
					sched.stopSign.Deal(SCHEDULER_CODE)
					return
				}
				sched.getReqChan() <- *temp
				remainder--
			}
			time.Sleep(interval)
		}
	}()
}
// 获取通道管理器持有的请求通道。
func (sched *myScheduler) getReqChan() chan base.Request {
	reqChan, err := sched.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}

// 获取通道管理器持有的响应通道。
func (sched *myScheduler) getRespChan() chan base.Response {
	respChan, err := sched.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

// 获取通道管理器持有的条目通道。
func (sched *myScheduler) getItemChan() chan base.Item {
	itemChan, err := sched.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}

// 获取通道管理器持有的错误通道。
func (sched *myScheduler) getErrorChan() chan error {
	errorChan, err := sched.chanman.ErrorChan()
	if err != nil {
		panic(err)
	}
	return errorChan
}
