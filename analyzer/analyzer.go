package analyzer

import (
	"webcrawler/base"
)

type Analyzer interface {
	Id() uint32
	Analyzer(respParsers []ParseResponse, resp *base.Response) ([]base.Data, []error) //根据规则分析响应并返回请求和条目
}
