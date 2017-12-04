package downloader

import (
	"webcrawler/base"
	"net/http"
	"webcrawler/middleware"
	"github.com/kataras/golog"
)

//id生成器
var downloaderIdGenerator middleware.IdGenerator = middleware.NewIdGenerator()

func genDownloaderId() uint32 {
	return downloaderIdGenerator.GetUint32()
}

type PageDownloader interface {
	Id() uint32                                        //获得id
	Download(req base.Request) (*base.Response, error) //根据请求下载网页并返回响应
}

type myPagedownloader struct {
	id     uint32       //id
	client *http.Client //http客户端
}

func NewPageDownloader(client *http.Client) PageDownloader {
	if client == nil {
		client = &http.Client{}
	}
	dl := &myPagedownloader{id: genDownloaderId(), client: client}
	return dl
}

func (dl *myPagedownloader) Id() uint32 {
	return dl.id
}

func (dl *myPagedownloader) Download(req base.Request) (*base.Response, error) {
	httpReq := req.HttpReq()
	golog.Infof("Do the request (url=%s)... \n", httpReq.URL)
	response, err := dl.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return base.NewResponse(response, req.Depth()), err
}
