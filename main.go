package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.forceup.in/zengliang/rpc2-center/common"
	"gitlab.forceup.in/zengliang/rpc2-center/logger"
	"gitlab.forceup.in/zengliang/rpc2-center/rpc"
	"gitlab.forceup.in/zengliang/rpc2-center/tools"
	"io/ioutil"
	"os"
	"time"
)

const centerCfg = "./res/center.json"
const nodeCfg = "./res/node.json"

var (
	Meta = ""
)

func main() {
	hh := tools.ParseMeta(Meta)
	bb, _ := json.Marshal(hh)
	fmt.Println(string(bb))

	logger.InitLogger(&logger.MyLogger{})

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	if cmd == "center" {
		err = runCenter()
	} else if cmd == "node" {
		err = runNode()
	} else {
		fmt.Println("need center or node command")
		return
	}

	fmt.Println("err:", err)
}

func runCenter() error {
	data, err := ioutil.ReadFile(centerCfg)
	if err != nil {
		return err
	}

	cfg := common.ConfigCenter{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	centerInst, err := rpc.NewCenter(cfg, Meta, func(reg *common.Register,
		status common.ConnectStatus) { })
	if err != nil {
		return err
	}

	apiGroup := centerInst.GetApiGroup()
	apiGroup.RegisterCaller("listsrv", func(req *common.Request, res *common.Response) {
		fmt.Println("call listsrv")
		res.Data.SetResult(centerInst.ListSrv())
	})
	apiGroup.RegisterCaller("ping", func(req *common.Request, res *common.Response) {
		fmt.Println("call ping")
		res.Data.SetResult("pong")
	})

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	rpc.StartCenter(ctx, centerInst)

	time.Sleep(time.Second * 1)
	for {
		fmt.Println("Input 'q' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "q" {
			cancel()
			break
		}
	}

	n.Info("Waiting all routine quit...")
	rpc.StopCenter(centerInst)
	c.Info("All routine is quit...")

	c.Info("wait 10 second to exit...")
	time.Sleep(time.Second*10)

	return nil
}

func runNode() error {
	data, err := ioutil.ReadFile(nodeCfg)
	if err != nil {
		return err
	}

	cfg := common.ConfigNode{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	nodeInst, err := rpc.NewNode(cfg, Meta, func(status common.ConnectStatus) { })
	if err != nil {
		return err
	}

	apiGroup := nodeInst.GetApiGroup()
	apiGroup.RegisterNotifier("update", func(req *common.Request) {
		fmt.Println("notify update")
	})
	apiGroup.RegisterCaller("ping", func(req *common.Request, res *common.Response) {
		fmt.Println("call ping")
		res.Data.SetResult(map[string]interface{}{
			"aa":map[string]string{
				"bb":"cc",
			},
		})
	})

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	rpc.StartNode(ctx, nodeInst)

	time.Sleep(time.Second * 1)
	for {
		fmt.Println("Input 'q' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "q" {
			cancel()
			break
		}
	}

	c.Info("Waiting all routine quit...")
	rpc.StopNode(nodeInst)
	c.Info("All routine is quit...")

	c.Info("wait 10 second to exit...")
	time.Sleep(time.Second*10)

	return nil
}
