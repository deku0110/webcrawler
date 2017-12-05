package analyzer

import (
	"webcrawler/base"
	"webcrawler/middleware"
	"errors"
	"net/url"
	"github.com/kataras/golog"
	"fmt"
)

var analyzerIdGenerator middleware.IdGenerator = middleware.NewIdGenerator()

func GenAnalyzerId() uint32 {
	return analyzerIdGenerator.GetUint32()
}

type Analyzer interface {
	Id() uint32
	Analyzer(respParsers []ParseResponse, resp *base.Response) ([]base.Data, []error) //根据规则分析响应并返回请求和条目
}

type myAnalyzer struct {
	id uint32
}

func NewAnalyzer() Analyzer {
	return &myAnalyzer{id:GenAnalyzerId()}
}

func (analyzer *myAnalyzer) Id() uint32 {
	return analyzer.id
}

func (analyzer *myAnalyzer) Analyzer(respParsers []ParseResponse, resp *base.Response) (dataList []base.Data, errorList []error) {
	if respParsers == nil {
		err := errors.New("The response Parser is invalid!")
		return nil,[]error{err}
	}
	httpResp := resp.HttpResp()
	if httpResp == nil {
		err := errors.New("The http response is invalid!")
		return nil,[]error{err}
	}
	var reqUrl *url.URL = httpResp.Request.URL
	golog.Infof("Parse the response (reqUrl=%s)...\n",reqUrl)
	respDepth := resp.Depth()
	//解析HTTP响应
	dataList = make([]base.Data,0)
	errorList = make([]error,0)
	
	for i,respParser := range respParsers {
		
		if respParser == nil {
			err := errors.New(fmt.Sprintf("The document parser [%d] is invalid!", i))
			errorList = append(errorList, err)
			continue
		}
		pDataList,pErrorList := respParser(httpResp, respDepth)
		if pDataList != nil {
			for _,pData := range pDataList {
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}

		if pErrorList != nil {
			for _,pError := range pErrorList {
				errorList = appendErrorList(errorList, pError)
			}
		}
	}
	return
}


// 添加请求值或条目值到列表。
func appendDataList(dataList []base.Data, data base.Data, respDepth uint32) []base.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*base.Request)
	//如果断言失败证明他不是个Request那么他一定是一个条目,直接追加到data中
	if !ok {
		return append(dataList, data)
	}
	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = base.NewRequest(req.HttpReq(), newDepth)
	}
	return append(dataList, req)
}

// 添加错误值到列表。
func appendErrorList(errorList []error, err error) []error {
	if err == nil {
		return errorList
	}
	return append(errorList, err)
}
