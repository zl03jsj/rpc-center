package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.forceup.in/zengliang/rpc2-center/common"
	"gitlab.forceup.in/zengliang/rpc2-center/loger"
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

	center, err := rpc.NewCenter(cfg, Meta, &loger.MyLoger{}, func(reg *common.Register,
		status common.ConnectStatus) { })
	if err != nil {
		return err
	}

	apiGroup := center.GetApiGroup()
	apiGroup.RegisterCaller("listsrv", func(req *common.Request, res *common.Response) {
		fmt.Println("call listsrv")
		res.Data.SetResult(center.ListSrv())
	})
	apiGroup.RegisterCaller("ping", func(req *common.Request, res *common.Response) {
		fmt.Println("call ping")
		res.Data.SetResult("pong")
	})

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	rpc.StartCenter(ctx, center)

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

	center.Info("Waiting all routine quit...")
	rpc.StopCenter(center)
	center.Info("All routine is quit...")

	center.Info("wait 10 second to exit...")
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

	node, err := rpc.NewNode(cfg, Meta, &loger.MyLoger{}, func(status common.ConnectStatus) { })
	if err != nil {
		return err
	}

	apiGroup := node.GetApiGroup()
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
	rpc.StartNode(ctx, node)

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

	node.Info("Waiting all routine quit...")
	rpc.StopNode(node)
	node.Info("All routine is quit...")

	node.Info("wait 10 second to exit...")
	time.Sleep(time.Second*10)

	return nil
}
