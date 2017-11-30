package middleware

import (
	"webcrawler/base"
	"sync"
	"fmt"
	"errors"
)

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
	Init(channelArgs base.ChannelArgs, reset bool) bool
	Close() bool
	ReqChan() (chan base.Request, error)
	RespChan() (chan base.Response, error)
	ItemChan() (chan base.Item, error)
	ErrorChan() (chan error, error)
	//获取通道管理器的状态
	Status() ChannelManagerStatus
	//获取摘要信息
	Summary() string
}

//通道管理器的实现类型
type myChannelManager struct {
	channelArgs base.ChannelArgs //通道的长度值
	reqCh       chan base.Request
	respCh      chan base.Response
	itemCh      chan base.Item
	errorCh     chan error
	status      ChannelManagerStatus //通道管理器的状态
	rwmutex     sync.RWMutex         //读写锁
}

func NewChannelManager(channelArgs base.ChannelArgs) ChannelManager {
	chanman := &myChannelManager{}
	chanman.Init(channelArgs, true)
	return chanman
}

func (cm *myChannelManager) Init(channelArgs base.ChannelArgs, reset bool) bool {
	if err := channelArgs.Check(); err != nil {
		panic(err)
	}
	cm.rwmutex.Lock()
	defer cm.rwmutex.Unlock()
	if cm.status == CHANNEL_MANAGER_STATUS_INITIALIZED && !reset {
		return false
	}
	cm.channelArgs = channelArgs
	cm.reqCh = make(chan base.Request, channelArgs.ReqChanLen())
	cm.respCh = make(chan base.Response, channelArgs.RespChanLen())
	cm.itemCh = make(chan base.Item, channelArgs.ItemChanLen())
	cm.errorCh = make(chan error, channelArgs.ErrorChanLen())
	cm.status = CHANNEL_MANAGER_STATUS_INITIALIZED
	return true
}

func (cm *myChannelManager) Close() bool {
	cm.rwmutex.Lock()
	defer cm.rwmutex.Unlock()
	if cm.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
		return false
	}
	close(cm.reqCh)
	close(cm.respCh)
	close(cm.itemCh)
	close(cm.errorCh)
	cm.status = CHANNEL_MANAGER_STATUS_CLOSED
	return true
}

func (cm *myChannelManager) ReqChan() (chan base.Request, error) {
	cm.rwmutex.Lock()
	defer cm.rwmutex.Unlock()
	if err := cm.checkStatus(); err != nil {
		return nil, err
	}
	return cm.reqCh, nil
}

func (cm *myChannelManager) RespChan() (chan base.Response, error) {
	cm.rwmutex.Lock()
	defer cm.rwmutex.Unlock()
	if err := cm.checkStatus(); err != nil {
		return nil, err
	}
	return cm.respCh, nil
}

func (cm *myChannelManager) ItemChan() (chan base.Item, error) {
	cm.rwmutex.Lock()
	defer cm.rwmutex.Unlock()
	if err := cm.checkStatus(); err != nil {
		return nil, err
	}
	return cm.itemCh,nil
}

func (cm *myChannelManager) ErrorChan() (chan error, error) {
	cm.rwmutex.Lock()
	defer cm.rwmutex.Unlock()
	if err := cm.checkStatus(); err != nil {
		return nil, err
	}
	return cm.errorCh,nil
}

func (cm *myChannelManager) Status() ChannelManagerStatus {
	return cm.status
}

var chanmanSummaryTemplate = "status: %s, " +
	"requestChannel: %d/%d, " +
	"responseChannel: %d/%d, " +
	"itemChannel: %d/%d, " +
	"errorChannel: %d/%d"

func (chanman *myChannelManager) Summary() string {
	summary := fmt.Sprintf(chanmanSummaryTemplate,
		statusNameMap[chanman.status],
		len(chanman.reqCh), cap(chanman.reqCh),
		len(chanman.respCh), cap(chanman.respCh),
		len(chanman.itemCh), cap(chanman.itemCh),
		len(chanman.errorCh), cap(chanman.errorCh))
	return summary
}
func (cm *myChannelManager) checkStatus() error {
	if cm.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	statusName, ok := statusNameMap[cm.status]
	if !ok {
		statusName = fmt.Sprintf("%d", cm.status)
	}
	errMsg := fmt.Sprintf("The undesirable status of channel manager:%s!\n", statusName)
	return errors.New(errMsg)
}
