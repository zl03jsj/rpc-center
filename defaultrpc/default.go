package defaultrpc

import (
	"gitlab.forceup.in/zengliang/rpc2-center/rpc"
)

var (
	defaultCenterInst *rpc.Center
	defaultNodeInst   *rpc.Node
)

func SetDefaultCenterInst(centerInst *rpc.Center) {
	defaultCenterInst = centerInst
}

func SetDefaultNodeInst(nodeInst *rpc.Node) {
	defaultNodeInst = nodeInst
}

func DefaultCenterInst() *rpc.Center {
	return defaultCenterInst
}

func DefaultNodeInst() *rpc.Node {
	return defaultNodeInst
}
