package rpc

import (
	"gitlab.forceup.in/Payment/rpc2-center/common"
	"fmt"
	"strings"
	"sync"
)

type (
	BeforApiCaller func(req *common.Request, res *common.Response) bool
	ApiCaller   func(req *common.Request, res *common.Response)
	ApiNotifier func(req *common.Request)

	ApiCallerInfo struct {
		Name    string
		Handler ApiCaller
	}

	ApiNotifierInfo struct {
		Name    string
		Handler ApiNotifier
	}

	ApiInfoGroup struct {
		apiCallerNameList []string
		apiCallerInfoMap  map[string]*ApiCallerInfo

		apiNotifierNameList []string
		apiNotifierInfoMap  map[string]*ApiNotifierInfo

		rWMutex sync.RWMutex
	}
)

func NewApiGroup() *ApiInfoGroup {
	ag := &ApiInfoGroup{
		apiCallerInfoMap:   make(map[string]*ApiCallerInfo),
		apiNotifierInfoMap: make(map[string]*ApiNotifierInfo),
	}

	return ag
}

func (ag *ApiInfoGroup) RegisterCaller(name string, handler ApiCaller) error {
	ag.rWMutex.Lock()
	defer ag.rWMutex.Unlock()

	name = strings.ToLower(name)
	if _, ok := ag.apiCallerInfoMap[name]; ok {
		return fmt.Errorf("caller name(%s) exist", name)
	}

	apiCallerInfo := &ApiCallerInfo{Handler: handler, Name: name}
	ag.apiCallerInfoMap[name] = apiCallerInfo
	ag.apiCallerNameList = append(ag.apiCallerNameList, name)

	return nil
}

func (ag *ApiInfoGroup) RegisterNotifier(name string, handler ApiNotifier) error {
	ag.rWMutex.Lock()
	defer ag.rWMutex.Unlock()

	name = strings.ToLower(name)
	if _, ok := ag.apiNotifierInfoMap[name]; ok {
		return fmt.Errorf("notifier name(%s) exist", name)
	}

	apiNotifierInfo := &ApiNotifierInfo{Handler: handler, Name: name}
	ag.apiNotifierInfoMap[name] = apiNotifierInfo
	ag.apiNotifierNameList = append(ag.apiNotifierNameList, name)

	return nil
}

func (ag *ApiInfoGroup) GetCallerNameList() []string {
	ag.rWMutex.RLock()
	defer ag.rWMutex.RUnlock()

	return ag.apiCallerNameList
}

func (ag *ApiInfoGroup) GetNotifierNameList() []string {
	ag.rWMutex.RLock()
	defer ag.rWMutex.RUnlock()

	return ag.apiNotifierNameList
}

func (ag *ApiInfoGroup) HandleCall(req *common.Request, res *common.Response) {
	ag.rWMutex.RLock()
	defer ag.rWMutex.RUnlock()

	h := ag.apiCallerInfoMap[strings.ToLower(req.Method.Function)]
	if h != nil {
		h.Handler(req, res)
	} else {
		res.Data.Err = common.ErrNotFindCaller
	}
}

func (ag *ApiInfoGroup) HandleNotify(req *common.Request, res *common.Response) {
	ag.rWMutex.RLock()
	defer ag.rWMutex.RUnlock()

	h := ag.apiNotifierInfoMap[strings.ToLower(req.Method.Function)]
	if h != nil {
		h.Handler(req)
	} else {
		res.Data.Err = common.ErrNotFindNotifier
	}
}
