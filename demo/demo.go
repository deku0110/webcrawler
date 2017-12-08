package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"webcrawler/analyzer"
	"webcrawler/base"
	"webcrawler/itempipeline"
	"webcrawler/scheduler"
	"webcrawler/tool"

	"github.com/PuerkitoBio/goquery"
	"github.com/kataras/golog"
)

//条目处理器
func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}
	//生成结果
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}

// 响应解析函数。只解析“A”标签。
func parseForATag(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	if httpResp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("Unsupported status code %d. (httpResponseCode=%v", httpResp))
		return nil, []error{err}
	}
	var reqUrl = httpResp.Request.URL
	var httpRespBody = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()
	dataList := make([]base.Data, 0)
	errs := make([]error, 0)
	//开始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}
	//查找A标签并提取链接地址
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		//暂不支持对javascript代码的解析
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}
			httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := base.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			imap := make(map[string]interface{})
			imap["parent_url"] = reqUrl
			imap["a.text"] = text
			imap["a.index"] = index
			item := base.Item(imap)
			dataList = append(dataList, &item)
		}
	})
	return dataList, errs
}

//获得响应及诶系函数的序列
func getResponseParsers() []analyzer.ParseResponse {
	parsers := []analyzer.ParseResponse{
		parseForATag,
	}
	return parsers
}
func genHttpClient() *http.Client {
	return &http.Client{}
}
func getItemProcessors() []itempipeline.ProcessItem {
	itemProcessors := []itempipeline.ProcessItem{
		processItem,
	}
	return itemProcessors
}

func record(level byte, content string) {
	if content == "" {
		return
	}
	switch level {
	case 0:
		golog.Info(content,"\n")
	case 1:
		golog.Warn(content,"\n")
	case 2:
		golog.Info(content,"\n")
	}
}

func main() {
	//创建调度器
	newScheduler := scheduler.NewScheduler()

	// 准备监控参数
	intervalNs := 10 * time.Millisecond
	maxIdleCount := uint(1000)
	// 开始监控
	checkCountChan := tool.Monitoring(newScheduler,
		intervalNs,
		maxIdleCount,
		true,
		false,
		record)

	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolSize := base.NewPoolBaseArgs(3, 3)
	crawDepth := uint32(1)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "http://www.sogo.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		golog.Error(err, "\n")
		return
	}
	// 开启调度器
	newScheduler.Start(channelArgs, poolSize, crawDepth, httpClientGenerator, respParsers, itemProcessors, firstHttpReq)

	<-checkCountChan
}
