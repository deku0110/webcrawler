package base

import "net/http"

type Data interface {
	Valid() bool //数据是否有效
}
//请求
type Request struct {
	httpReq *http.Request //HTTP请求的指针值
	depth   uint32        //请求的深度
}

func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

func (req *Request) HttpReq() *http.Request {
	return req.httpReq
}
func (req *Request) Depth() uint32 {
	return req.depth
}

func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}

//响应
type Response struct {
	httpResp *http.Response //响应
	depth    uint32         //请求的深度
}

func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

func (resp *Response) HttpResp() *http.Response {
	return resp.httpResp
}

func (resp *Response) Depth() uint32 {
	return resp.depth
}

func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}
//条目
type Item map[string]interface{}

func (item Item) Valid() bool {
	return item != nil
}
