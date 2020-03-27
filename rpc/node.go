package rpc

import (
	"context"
	"fmt"
	"github.com/henly2/rpc2"
	"gitlab.forceup.in/zengliang/rpc2-center/common"
	"gitlab.forceup.in/zengliang/rpc2-center/l4g"
	"gitlab.forceup.in/zengliang/rpc2-center/tools"
	"net"
	"sync"
	"time"
)

type (
	ConnectCenterStatusCallBack func(status common.ConnectStatus)
	Node                        struct {
		*rpc2.Client
		*l4g.ILoger
		rwMu sync.RWMutex

		cfgNode common.ConfigNode
		cb      ConnectCenterStatusCallBack

		wg sync.WaitGroup

		apiGroup *ApiInfoGroup

		regData common.Register

		statusMu sync.Mutex
		stopped  bool

		befor_bycall BeforApiCaller
	}
)

func NewNode(conf common.ConfigNode, meta string, cb ConnectCenterStatusCallBack) (*Node, error) {
	node := &Node{
		cfgNode:  conf,
		cb:       cb,
		apiGroup: NewApiGroup(),
	}

	node.regData.StartAt = tools.GetDateNowString()
	node.regData.Meta = tools.ParseMeta(meta)
	node.regData.Env = tools.GetOsEnv(node.cfgNode.Env)
	node.regData.Service = node.cfgNode.Service
	node.ILoger = l4g.GetL4g("rpc2-client")

	return node, nil
}

func (n *Node) GetApiGroup() *ApiInfoGroup {
	return n.apiGroup
}

func StartNode(ctx context.Context, n *Node) {
	n.initFunction()

	n.startToCenter(ctx)
}

func StopNode(n *Node) {
	n.setStopped(true)
	n.wg.Wait()
}

func (n *Node) setStopped(status bool) {
	n.statusMu.Lock()
	defer n.statusMu.Unlock()
	n.stopped = status
}

func (n *Node) isStopped() bool {
	n.statusMu.Lock()
	defer n.statusMu.Unlock()
	return n.stopped
}

func (n *Node) initFunction() {
	n.regData.CallerList = n.apiGroup.GetCallerNameList()
	n.regData.NotifierList = n.apiGroup.GetNotifierNameList()
}

func (n *Node) SetBeforCall(befor_call BeforApiCaller) {
	n.befor_bycall = befor_call
}

func (n *Node) byCall(client *rpc2.Client, req *common.Request, res *common.Response) error {
	n.Info("begin call:%s", req.Method.Function)
	defer n.Info("end call:%s-%d", req.Method.Function, res.Data.Err)

	if n.apiGroup == nil {
		res.Data.Err = common.ErrInternal
		return nil
	}

	if n.befor_bycall != nil {
		if !n.befor_bycall(req, res) {
			return nil
		}
	}

	n.apiGroup.HandleCall(req, res)

	return nil
}

func (n *Node) byNotify(client *rpc2.Client, req *common.Request, res *common.Response) error {
	n.Info("begin notify:%s", req.Method.Function)
	defer n.Info("end notify:%s-%d", req.Method.Function, res.Data.Err)

	if n.apiGroup == nil {
		res.Data.Err = common.ErrInternal
		return nil
	}

	n.apiGroup.HandleNotify(req, res)

	return nil
}

func (n *Node) byKeepAlive(client *rpc2.Client, req *string, res *string) error {
	n.Debug("begin keepalive")
	defer n.Debug("end keepalive")

	if *req != "ping" {
		*res = *req
	} else {
		*res = "pong"
	}

	return nil
}

func (n *Node) Call(req *common.Request, res *common.Response) error {
	if n.isStopped() {
		return fmt.Errorf("client is stopped")
	}

	n.rwMu.RLock()
	defer n.rwMu.RUnlock()

	var err error
	if n.Client != nil {
		err = n.Client.Call(common.MethodCenterCall, req, res)
	} else {
		err = fmt.Errorf("client is nil")
	}

	return err
}

func (n *Node) Notify(req *common.Request, res *common.Response) error {
	if n.isStopped() {
		return fmt.Errorf("client is stopped")
	}

	n.rwMu.RLock()
	defer n.rwMu.RUnlock()

	var err error
	if n.Client != nil {
		err = n.Client.Notify(common.MethodCenterNotify, req)
	} else {
		err = fmt.Errorf("client is nil")
	}
	return err
}

func (n *Node) connectToCenter() (*rpc2.Client, error) {
	conn, err := net.Dial("tcp", n.cfgNode.RpcAddr)
	if err != nil {
		return nil, err
	}

	clt := rpc2.NewClient(conn)
	return clt, nil
}

func (n *Node) registerToCenter() error {
	var res string

	var err error
	if n.Client != nil {
		err = n.Client.Call(common.MethodCenterRegister, &n.regData, &res)
		n.Info("Register to center ok %s.%s", n.cfgNode.Version, n.cfgNode.Name)
	} else {
		err = fmt.Errorf("client is nil")
	}
	return err
}

func (n *Node) unRegisterToCenter() error {
	var res string

	var err error
	if n.Client != nil {
		n.Client.Call(common.MethodCenterUnRegister, n.cfgNode, &res)
		n.Info("UnRegister to center ok %s.%s", n.cfgNode.Version, n.cfgNode.Name)
	} else {
		err = fmt.Errorf("client is nil")
	}

	return err
}

func (n *Node) startToCenter(ctx context.Context) {
	go func() {
		n.wg.Add(1)
		n.Info("startToCenter loop start...")

		defer func() {
			n.wg.Done()
			n.Info("startToCenter loop stop...")
		}()

		for {
			// connect and register
			func() {
				n.rwMu.Lock()
				defer n.rwMu.Unlock()

				var err error
				if n.Client == nil {
					n.Info("client try to connect...")
					n.Client, err = n.connectToCenter()
					if n.Client != nil && err == nil {
						n.Info("client connect to center...")
						n.Client.Handle(common.MethodNodeCall, n.byCall)
						n.Client.Handle(common.MethodNodeNotify, n.byNotify)
						n.Client.Handle(common.MethodNodeKeepAlive, n.byKeepAlive)

						go n.Client.Run()

						n.registerToCenter()

						if n.cb != nil {
							n.cb(common.ConnectStatusConnected)
						}
					}
				}

				if err != nil {
					if n.Client != nil {
						n.Client.Close()
						n.Client = nil
					}
					n.Error("connect failed, %s", err.Error())
				}
			}()

			// listen
			func() {
				n.rwMu.RLock()
				defer n.rwMu.RUnlock()

				if n.Client == nil {
					return
				}

				n.Info("client run...")
				select {
				case <-ctx.Done():
					n.Error("user disconnect client ...")
					break
				case <-n.Client.DisconnectNotify():
					n.Error("client disconnect...")
					break
				}

				if n.cb != nil {
					n.cb(common.ConnectStatusDisConnected)
				}
			}()

			// unregister and close
			func() {
				n.rwMu.Lock()
				defer n.rwMu.Unlock()
				n.unRegisterToCenter()
				if n.Client != nil {
					n.Client.Close()
					n.Client = nil

					n.Info("reset client...")
				}
			}()

			if n.isStopped() {
				return
			}

			n.Info("wait 5 second to connect...")
			time.Sleep(time.Second * 5)
		}
	}()
}
