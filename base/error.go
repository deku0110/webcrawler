package base

import (
	"bytes"
	"fmt"
)

const (
	DOWNLOADER_ERROR     ErrorType = "Downloader Error"
	ANALYZER_ERROR       ErrorType = "Analyzer Error"
	ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

type CrawlerError interface {
	Type() ErrorType //获得错误类型
	Error() string   //获得错误提示信息
}

type ErrorType string

//爬虫错误的实现
type myCrawlerError struct {
	errType    ErrorType //错误类型
	errMsg     string    //错误信息
	fullErrMsg string    //完整错误信息
}

func NewCrawLerError(errType ErrorType,errMsg string) CrawlerError {
	return &myCrawlerError{errType:errType,errMsg:errMsg}
}

func (ce *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error:")
	if ce.errType != "" {
		buffer.WriteString(string(ce.errType))
		buffer.WriteString(":")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrMsg = fmt.Sprintf("%s\n",buffer.String())
	return
}
func (ce *myCrawlerError) Type() ErrorType {
	return ce.errType
}

func (ce *myCrawlerError) Error() string {
	if ce.fullErrMsg == "" {
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}
