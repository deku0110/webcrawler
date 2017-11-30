package scheduler

import (
	"webcrawler/base"
	"net/http"
	"webcrawler/analyzer"
	"webcrawler/itempipeline"
)

//用来生成httpClient的方法
type GenHttpClient func() *http.Client

type Scheduler interface {
	Start(channelArgs base.ChannelArgs,poolBaseArgs base.PoolBaseArgs,crawDepth uint32,
		httpClientGenerator GenHttpClient,respParsers []analyzer.ParseResponse,
			item []itempipeline.ProcessItem,firstHttpReq *http.Request) error

	Stop() bool

	Running() bool
	//错误通道,若为nil 表示通道不可用或者调度器处于停止状态
	ErrorChan() <-chan error
	//判断所有模块是否都处于空闲状态
	Idle() bool
	//摘要信息
	Summary(prefix string) SchedSummary
}
