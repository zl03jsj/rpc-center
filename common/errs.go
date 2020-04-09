package common

import (
	"sync"
)

type ErrCode int64

const (
	ErrOk = ErrCode(iota)

	ErrInternal        = ErrCode(1000) // 内部错误
	ErrNotFindService  = ErrCode(1001) // 没有找到服务
	ErrCallFailed      = ErrCode(1002) // 调用失败
	ErrNotFindCaller   = ErrCode(1003) // 没有找到方法
	ErrNotFindNotifier = ErrCode(1004) // 没有找到通知
	ErrDataCorrupted   = ErrCode(1005) // 数据损坏
)

var err_msgs = map[ErrCode]string{
	ErrOk:              "success",
	ErrInternal:        "internal error",
	ErrNotFindService:  "service not found",
	ErrCallFailed:      "call method failed",
	ErrNotFindCaller:   "method not found",
	ErrNotFindNotifier: "notifier not found",
	ErrDataCorrupted:   "invalid data"}

var mutx sync.Mutex

func RegistErrorInfo(code ErrCode, msg string) bool {
	mutx.Lock()
	defer mutx.Unlock()
	if _, exist := err_msgs[code]; exist {
		return false
	}
	err_msgs[code] = msg
	return true
}

func (self ErrCode) String() string {
	msg, isok := err_msgs[self]
	if !isok {
		return "unkown error"
	}
	return msg
}
