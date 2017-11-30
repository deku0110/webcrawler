package middleware

import "webcrawler/base"

// 被用来表示通道管理器的状态的类型。
type ChannelManagerStatus uint8

const (
	CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManagerStatus = 0 // 未初始化状态。
	CHANNEL_MANAGER_STATUS_INITIALIZED   ChannelManagerStatus = 1 // 已初始化状态。
	CHANNEL_MANAGER_STATUS_CLOSED        ChannelManagerStatus = 2 // 已关闭状态。
)

// 表示状态代码与状态名称之间的映射关系的字典。
var statusNameMap = map[ChannelManagerStatus]string{
	CHANNEL_MANAGER_STATUS_UNINITIALIZED: "uninitialized",
	CHANNEL_MANAGER_STATUS_INITIALIZED:   "initialized",
	CHANNEL_MANAGER_STATUS_CLOSED:        "closed",
}
type ChannelManager interface {
	//reset代表是否重新初始化通道管理器
	Init(channelArgs base.ChannelArgs,reset bool) bool

	Close() bool
	ReqChan() (chan base.Request,error)
	RespChan() (chan base.Response,error)
	ItemChan() (chan base.Item,error)
	ErrorChan() (chan error,error)
	//获取通道管理器的状态
	Status() ChannelManagerStatus
	//获取摘要信息
	Summary() string
}
