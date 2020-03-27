package common

const (
	ErrOk = iota

	ErrInternal        = 1000 // 内部错误
	ErrNotFindService  = 1001 // 没有找到服务
	ErrCallFailed      = 1002 // 调用失败
	ErrNotFindCaller   = 1003 // 没有找到方法
	ErrNotFindNotifier = 1004 // 没有找到通知
	ErrDataCorrupted   = 1005 // 数据损坏
)
