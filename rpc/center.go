package rpc

import (
	"context"
	"encoding/base64"
	"github.com/henly2/rpc2"
	"gitlab.forceup.in/zengliang/rpc2-center/common"
	"gitlab.forceup.in/zengliang/rpc2-center/httpserver"
	"gitlab.forceup.in/zengliang/rpc2-center/logger"
	"gitlab.forceup.in/zengliang/rpc2-center/tools"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type (
	NodeConnectStatusCallBack func(reg *common.Register, status common.ConnectStatus)
	Center                    struct {
		*rpc2.Server

		cfgCenter common.ConfigCenter
		cb        NodeConnectStatusCallBack

		rwMu                sync.RWMutex
		verNameMapNodeGroup map[string]*NodeGroup
		clientMapNodeGroup  map[*rpc2.Client]*NodeGroup

		wg sync.WaitGroup

		apiGroup *ApiInfoGroup

		httpServer *httpserver.HttpServer

		regData common.Register

	}
)

func NewCenter(conf common.ConfigCenter, meta string, cb NodeConnectStatusCallBack) (*Center, error) {
	center := &Center{
		cfgCenter:           conf,
		cb:                  cb,
		verNameMapNodeGroup: make(map[string]*NodeGroup),
		clientMapNodeGroup:  make(map[*rpc2.Client]*NodeGroup),
		apiGroup:            NewApiGroup(),
		httpServer:          httpserver.NewHttpServer(),
	}

	center.regData.StartAt = tools.GetDateNowString()
	center.regData.Meta = tools.ParseMeta(meta)
	center.regData.Env = tools.GetOsEnv(center.cfgCenter.Env)
	center.regData.Service = center.cfgCenter.Service

	// rpc2
	center.Server = rpc2.NewServer()

	return center, nil
}

func (c *Center) GetApiGroup() *ApiInfoGroup {
	return c.apiGroup
}

func StartCenter(ctx context.Context, c *Center) {
	c.initFunction()

	c.startHttpServer(ctx)

	c.startTcpServer(ctx)

	if c.cfgCenter.KeepAlive > 0 {
		c.startLoopKeepAlive(ctx)
	}
}

func StopCenter(c *Center) {
	c.wg.Wait()

	c.httpServer.Stop()
}

func (c *Center) initFunction() {
	// register
	var res string
	c.regData.CallerList = c.apiGroup.GetCallerNameList()
	c.regData.NotifierList = c.apiGroup.GetNotifierNameList()
	c.byRegister(nil, &c.regData, &res)
}

func (c *Center) byRegister(client *rpc2.Client, reg *common.Register, res *string) error {
	logger.GlobalLogger.Info("register client %s", reg.GetInstance())

	err := func() error {
		c.rwMu.Lock()
		defer c.rwMu.Unlock()

		srvKey := reg.Service.GetKey()

		nodeGroup, ok := c.verNameMapNodeGroup[srvKey]
		if !ok {
			nodeGroup = &NodeGroup{}
			c.verNameMapNodeGroup[srvKey] = nodeGroup
		}

		c.clientMapNodeGroup[client] = nodeGroup

		err := nodeGroup.Register(client, reg)
		if err != nil {
			*res = "failed"
			logger.GlobalLogger.Error("register %s err: %s", reg.GetInstance(), err.Error())
			return err
		}
		*res = "ok"

		logger.GlobalLogger.Info("register done: all group=%d, all clients=%d", len(c.verNameMapNodeGroup), len(c.clientMapNodeGroup))
		return nil
	}()

	if err == nil {
		if c.cb != nil {
			c.cb(reg, common.ConnectStatusConnected)
		}
	}

	return err
}

func (c *Center) byUnRegister(client *rpc2.Client, reg *string, res *string) error {
	c.disconnectClient(client)
	*res = "ok"
	return nil
}

func (c *Center) Call(req *common.Request) *common.Response {
	var res = &common.Response{}
	c.byCall(nil, req, res)
	return res
}

func (c *Center) byCall(fromClient *rpc2.Client, req *common.Request, res *common.Response) error {
	c.wg.Add(1)
	defer c.wg.Done()

	logger.GlobalLogger.Debug("by call %s:%s", req.Method.GetInstance(), req.Method.Function)

	c.callFunction(fromClient, req, res)

	return nil
}

func (c *Center) byNotify(fromClient *rpc2.Client, req *common.Request, res *common.Response) error {
	c.wg.Add(1)
	defer c.wg.Done()

	logger.GlobalLogger.Debug("by notify %s:%s", req.Method.GetInstance(), req.Method.Function)

	c.notifyFunction(fromClient, req, res)

	return nil
}

func (c *Center) startHttpServer(ctx context.Context) {
	// http
	logger.GlobalLogger.Info("Start http server on %s", c.cfgCenter.HttpPort)

	c.httpServer.RegisterHandler("/call/", http.HandlerFunc(c.handleCall))
	c.httpServer.RegisterHandler("/notify/", http.HandlerFunc(c.handleNotify))

	c.httpServer.Start(c.cfgCenter.HttpPort)
}

func (c *Center) disconnectClient(client *rpc2.Client) {
	if client != nil {
		client.Close()
	}

	reg := func()*common.Register{
		c.rwMu.Lock()
		defer c.rwMu.Unlock()

		nodeGroup, ok := c.clientMapNodeGroup[client]
		if nodeGroup == nil || !ok {
			return nil
		}

		reg, _ := nodeGroup.UnRegister(client)

		if nodeGroup.GetNodeCount() == 0 {
			srvInfo := nodeGroup.GetNodeInfo()
			verName := strings.ToLower(srvInfo.GetKey())
			delete(c.verNameMapNodeGroup, verName)
		}

		delete(c.clientMapNodeGroup, client)

		return reg
	}()

	if reg != nil {
		if c.cb != nil {
			c.cb(reg, common.ConnectStatusDisConnected)
		}
	}
}

func (c *Center) startTcpServer(ctx context.Context) {
	c.Server.OnConnect(func(client *rpc2.Client) {
		logger.GlobalLogger.Info("rpc2 client connect...")
	})

	c.Server.OnDisconnect(func(client *rpc2.Client) {
		logger.GlobalLogger.Info("rpc2 client disconnect...")

		c.disconnectClient(client)
	})

	c.Server.Handle(common.MethodCenterRegister, c.byRegister)
	c.Server.Handle(common.MethodCenterUnRegister, c.byUnRegister)
	c.Server.Handle(common.MethodCenterCall, c.byCall)
	c.Server.Handle(common.MethodCenterNotify, c.byNotify)

	logger.GlobalLogger.Info("Start RPC Tcp server on %s", c.cfgCenter.RpcPort)

	addr, err := net.ResolveTCPAddr("tcp", c.cfgCenter.RpcPort)
	if err != nil {
		logger.GlobalLogger.Fatal("%s", err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logger.GlobalLogger.Fatal("%s", err)
	}

	go func() {
		c.wg.Add(1)
		defer c.wg.Done()

		logger.GlobalLogger.Info("Tcp server routine running... ")

		go c.Server.Accept(listener)
		<-ctx.Done()

		logger.GlobalLogger.Info("Tcp server routine stopped... ")
	}()
}

func (c *Center) startLoopKeepAlive(ctx context.Context) {
	logger.GlobalLogger.Trace("start keep alive loop...")

	go func() {
		c.wg.Add(1)
		logger.GlobalLogger.Trace("keep alive loop running...")

		defer func() {
			c.wg.Done()
			logger.GlobalLogger.Trace("keep alive loop exit...")
		}()

		for {
			select {
				case <-ctx.Done():
					return
			default:
				time.Sleep(time.Duration(c.cfgCenter.KeepAlive) * time.Second)
			}

			logger.GlobalLogger.Trace("doing keep alive...")
			cc := c.callKeepAlive()
			if len(cc) != 0 {
				for _, client := range cc {
					c.disconnectClient(client)
				}
			}
			logger.GlobalLogger.Trace("doing keep alive done...%d", len(cc))
		}
	}()
}

func (c *Center)callKeepAlive() []*rpc2.Client {
	c.rwMu.RLock()
	defer c.rwMu.RUnlock()

	cc := []*rpc2.Client{}
	for client, _ := range c.clientMapNodeGroup {
		if client == nil{
			continue
		}

		res := ""
		err := client.Call(common.MethodNodeKeepAlive, "ping", &res)
		if err == nil && res == "pong" {
			continue
		}

		cc = append(cc, client)
		if err != nil {
			logger.GlobalLogger.Error("keepalive err: %s", err.Error())
			continue
		}
		if res != "pong" {
			logger.GlobalLogger.Error("keepalive err: %s", "not response pong")
			continue
		}
	}

	return cc
}

//  call a srv node
func (c *Center) callFunction(fromClient *rpc2.Client, req *common.Request, res *common.Response) {
	srvKey := strings.ToLower(req.Method.GetKey())
	logger.GlobalLogger.Trace("call %s:%s", req.Method.GetInstance(), req.Method.Function)
	defer logger.GlobalLogger.Trace("call %s:%s ret=%d",
		req.Method.GetInstance(), req.Method.Function, res.Data.Err)

	c.rwMu.RLock()
	defer c.rwMu.RUnlock()

	if srvKey == c.cfgCenter.GetKey() {
		if c.apiGroup == nil {
			res.Data.Err = common.ErrInternal
			return
		}

		c.apiGroup.HandleCall(req, res)
		return
	}

	if srvNodeGroup, ok := c.verNameMapNodeGroup[srvKey]; ok {
		srvNodeGroup.Call(fromClient, req, res)
		return
	}

	res.Data.Err = common.ErrNotFindService
	return
}

//  notify a srv node
func (c *Center) notifyFunction(fromClient *rpc2.Client, req *common.Request, res *common.Response) {
	srvKey := strings.ToLower(req.Method.GetKey())
	logger.GlobalLogger.Trace("notify %s:%s", req.Method.GetInstance(), req.Method.Function)
	defer logger.GlobalLogger.Trace("notify %s:%s ret=%d", req.Method.GetInstance(), req.Method.Function, res.Data.Err)

	c.rwMu.RLock()
	defer c.rwMu.RUnlock()

	if srvKey == c.cfgCenter.GetKey() {
		if c.apiGroup == nil {
			res.Data.Err = common.ErrInternal
			return
		}

		c.apiGroup.HandleNotify(req, res)
		return
	}

	if srvNodeGroup, ok := c.verNameMapNodeGroup[srvKey]; ok {
		srvNodeGroup.Notify(fromClient, req, res)
		return
	}

	res.Data.Err = common.ErrNotFindService
	return
}

func (c *Center) handleCall(w http.ResponseWriter, req *http.Request) {
	logger.GlobalLogger.Trace("Http server Accept a call client: %s", req.RemoteAddr)
	defer req.Body.Close()

	//w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	//w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	c.wg.Add(1)
	defer c.wg.Done()

	userResponse := common.HttpUserResponse{}
	func() {
		//fmt.Println("path=", req.URL.Path)
		reqData := common.Request{}
		reqData.Method.FromPath(req.URL.Path)
		reqData.Method.Tag = req.URL.Query().Get("tag")

		// get argv
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.GlobalLogger.Error("call http handler: %s", err.Error())
			userResponse.Err = common.ErrDataCorrupted
			return
		}

		reqData.Data.Value = base64.StdEncoding.EncodeToString(b)

		resData := common.Response{}
		c.callFunction(nil, &reqData, &resData)

		if resData.Data.Err != common.ErrOk {
			logger.GlobalLogger.Error("call http handler: %d", resData.Data.Err)
			userResponse.Err = resData.Data.Err
			userResponse.ErrMsg = resData.Data.ErrMsg
			return
		}

		err = resData.Data.GetResult(&userResponse.Result)
		if err != nil {
			logger.GlobalLogger.Error("call http handler: %s", err.Error())
			userResponse.Err = common.ErrDataCorrupted
			return
		}
	}()

	if userResponse.Err != common.ErrOk {
		logger.GlobalLogger.Error("handleCall request err: %d-%s", userResponse.Err, userResponse.ErrMsg)
	}

	// write back http
	connectionType := req.Header.Get("Connection")
	w.Header().Set("Connection", connectionType)
	w.Header().Set("Content-Type", "application/json")

	httpserver.ResponseDataByIndent(w, userResponse)
	return
}

func (c *Center) handleNotify(w http.ResponseWriter, req *http.Request) {
	logger.GlobalLogger.Debug("Http server Accept a notify client: %s", req.RemoteAddr)
	defer req.Body.Close()

	//w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	//w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型

	c.wg.Add(1)
	defer c.wg.Done()

	userResponse := common.UserResponse{}
	func() {
		//fmt.Println("path=", req.URL.Path)
		reqData := common.Request{}
		reqData.Method.FromPath(req.URL.Path)
		reqData.Method.Tag = req.URL.Query().Get("tag")

		// get argv
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.GlobalLogger.Error("notify http handler: %s", err.Error())
			userResponse.Err = common.ErrDataCorrupted
			return
		}

		reqData.Data.Value = base64.StdEncoding.EncodeToString(b)

		resData := common.Response{}
		c.notifyFunction(nil, &reqData, &resData)

		if resData.Data.Err != common.ErrOk {
			logger.GlobalLogger.Error("notify http handler: %d", resData.Data.Err)
			userResponse.Err = resData.Data.Err
			userResponse.ErrMsg = resData.Data.ErrMsg
			return
		}

		err = resData.Data.GetResult(&userResponse.Result)
		if err != nil {
			logger.GlobalLogger.Error("notify http handler: %s", err.Error())
			userResponse.Err = common.ErrDataCorrupted
			return
		}
	}()

	if userResponse.Err != common.ErrOk {
		logger.GlobalLogger.Error("handleNotify request err: %d-%s", userResponse.Err, userResponse.ErrMsg)
	}

	// write back http
	connectionType := req.Header.Get("Connection")
	w.Header().Set("Connection", connectionType)
	w.Header().Set("Content-Type", "application/json")

	httpserver.ResponseDataByIndent(w, userResponse)
	return
}

func (c *Center) ListSrv() map[string][]common.Register {
	c.rwMu.RLock()
	defer c.rwMu.RUnlock()

	srvInfoList := make(map[string][]common.Register)
	for srvKey, v := range c.verNameMapNodeGroup {
		srvInfoNodes := v.GetNodes()

		srvInfoList[srvKey] = srvInfoNodes
	}

	return srvInfoList
}

func (c *Center) Name() string {
	return c.cfgCenter.Name
}
