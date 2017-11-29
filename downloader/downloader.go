package downloader

import "webcrawler/base"

type PageDownloader interface {
	ID() int32                                         //获得id
	Download(req base.Request) (*base.Response, error) //根据请求下载网页并返回响应
}
