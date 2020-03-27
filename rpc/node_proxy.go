package rpc

import (
	"github.com/zl03jsj/log4go"
	"gitlab.forceup.in/zengliang/rpc2-center/common"
	"github.com/henly2/rpc2"
	"strings"
	"sync"
	"sync/atomic"
)

type (
	NodeInfo struct {
		client *rpc2.Client

		RegisterData common.Register
	}

	NodeGroup struct {
		log4go.Logger
		nodeInfo common.Service

		callFunctionMap   map[string]interface{}
		notifyFunctionMap map[string]interface{}

		rwMu  sync.RWMutex
		index int64
		nodes []*NodeInfo
	}
)

func (sng *NodeGroup) Register(client *rpc2.Client, reg *common.Register) error {
	sng.rwMu.Lock()
	defer sng.rwMu.Unlock()

	sng.nodeInfo.Version = reg.Version
	sng.nodeInfo.Name = reg.Name

	if sng.callFunctionMap == nil {
		sng.callFunctionMap = make(map[string]interface{})
	}
	if sng.notifyFunctionMap == nil {
		sng.notifyFunctionMap = make(map[string]interface{})
	}
	for _, cc := range reg.CallerList {
		sng.callFunctionMap[strings.ToLower(cc)] = struct{}{}
	}
	for _, cc := range reg.NotifierList {
		sng.notifyFunctionMap[strings.ToLower(cc)] = struct{}{}
	}

	si := &NodeInfo{
		client:       client,
		RegisterData: *reg,
	}
	sng.nodes = append(sng.nodes, si)

	sng.Debug("reg-%s.%s(%s), all-%d", reg.Version, reg.Name, reg.Tag, len(sng.nodes))
	return nil
}

func (sng *NodeGroup) UnRegister(client *rpc2.Client) (*common.Register, error) {
	sng.rwMu.Lock()
	defer sng.rwMu.Unlock()

	reg := &common.Register{}
	for i, v := range sng.nodes {
		if v.client == client {
			*reg = v.RegisterData
			sng.nodes = append(sng.nodes[:i], sng.nodes[i+1:]...)
			break
		}
	}

	sng.Debug("unreg-%s.%s(%s), all-%d", sng.nodeInfo.Version, sng.nodeInfo.Name, reg.Tag, len(sng.nodes))
	return reg, nil
}

func (sng *NodeGroup) GetNodeInfo() common.Service {
	sng.rwMu.RLock()
	defer sng.rwMu.RUnlock()

	return sng.nodeInfo
}

func (sng *NodeGroup) GetNodes() []common.Register {
	sng.rwMu.RLock()
	defer sng.rwMu.RUnlock()

	infos := []common.Register{}
	for _, v := range sng.nodes {
		infos = append(infos, v.RegisterData)
	}

	return infos
}

func (sng *NodeGroup) GetNodeCount() int {
	sng.rwMu.RLock()
	defer sng.rwMu.RUnlock()

	return len(sng.nodes)
}

type futureReciveRes chan *rpc2.Call

func make_faild_futrueRes(res *common.Response, err_code int) futureReciveRes {
	receive := make(chan *rpc2.Call, 1)
	res.Data.Err = err_code
	receive <- &rpc2.Call{}
	return receive
}

func (r futureReciveRes) Done() {
	call := <-r
	response, isok := call.Reply.(*common.Response)
	if !isok {
		return
	}

	if call.Error != nil {
		response.Data.Err = common.ErrCallFailed
		response.Data.ErrMsg = call.Error.Error()
	}
}

func (sng *NodeGroup) Go(fromClient *rpc2.Client, req *common.Request,
	res *common.Response) futureReciveRes {
	sng.rwMu.RLock()
	defer sng.rwMu.RUnlock()

	if _, ok := sng.callFunctionMap[req.Method.Function]; !ok {
		return make_faild_futrueRes(res, common.ErrNotFindCaller)
	}

	node := sng.getCallTagNode(fromClient, req.Method.Tag)
	if node == nil {
		return make_faild_futrueRes(res, common.ErrNotFindService)
	}

	return node.client.Go(common.MethodNodeCall, req, res,
		make(chan *rpc2.Call, 1)).Done
}

func (sng *NodeGroup) Call2(fromClient *rpc2.Client,
	req *common.Request, res *common.Response) {
	var receiver = sng.Go(fromClient, req, res)
	receiver.Done()
}

func (sng *NodeGroup) Call(fromClient *rpc2.Client, req *common.Request, res *common.Response) {
	sng.rwMu.RLock()
	defer sng.rwMu.RUnlock()

	if _, ok := sng.callFunctionMap[strings.ToLower(req.Method.Function)]; !ok {
		res.Data.Err = common.ErrNotFindCaller
		return
	}

	node := sng.getCallTagNode(fromClient, req.Method.Tag)
	if node == nil {
		res.Data.Err = common.ErrNotFindService
		return
	}

	err := node.client.Call(common.MethodNodeCall, req, res)
	if err != nil {
		sng.Error("#Call %s:%s srv:%s", req.Method.GetInstance(), req.Method.Function, err.Error())

		res.Data.Err = common.ErrCallFailed
		return
	}
}

func (sng *NodeGroup) Notify(client *rpc2.Client, req *common.Request, res *common.Response) {
	sng.rwMu.RLock()
	defer sng.rwMu.RUnlock()

	if _, ok := sng.notifyFunctionMap[req.Method.Function]; !ok {
		res.Data.Err = common.ErrNotFindNotifier
		return
	}

	for _, node := range sng.nodes {
		if node != nil && node.client != client {
			if req.Method.Tag == "" || req.Method.Tag == node.RegisterData.Tag {
				err := node.client.Notify(common.MethodNodeNotify, req)
				if err != nil {
					sng.Error("#Notify %s:%s srv:%s", req.Method.GetInstance(), req.Method.Function, err.Error())
					res.Data.Err = common.ErrCallFailed
				}
			}
		}
	}
}

func (sng *NodeGroup) getCallTagNode(fromClient *rpc2.Client, tag string) *NodeInfo {
	length := int64(len(sng.nodes))
	if length == 0 {
		return nil
	}

	if tag == "" {
		for i := int64(0); i < length; i++ {
			atomic.AddInt64(&sng.index, 1)
			atomic.CompareAndSwapInt64(&sng.index, length, 0)
			index := sng.index % length
			if sng.nodes[index].client != fromClient {
				return sng.nodes[index]
			}
		}
	} else {
		for _, node := range sng.nodes {
			if node.client != fromClient && node.RegisterData.Tag == tag {
				return node
			}
		}
	}

	return nil
}
