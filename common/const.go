package common

import (
	"fmt"
)

const (
	MethodCenterRegister   = "Center.Register"
	MethodCenterUnRegister = "Center.UnRegister"
	MethodCenterCall       = "Center.Call"
	MethodCenterNotify     = "Center.Notify"

	MethodNodeCall      = "Node.Call"
	MethodNodeNotify    = "Node.Notify"
	MethodNodeKeepAlive = "Node.KeepAlive"
)

type ConnectStatus int

var statusStrings = map[ConnectStatus]string{
	ConnectStatusConnected:    "connected",
	ConnectStatusDisConnected: "disconnected",
}

func (cs ConnectStatus) String() string {
	if s, isOk := statusStrings[cs]; isOk {
		return s
	}

	return fmt.Sprintf("unkown status:%d", cs)
}

func (cs ConnectStatus) Int() int {
	return int(cs)
}

const (
	ConnectStatusConnected    = ConnectStatus(1)
	ConnectStatusDisConnected = ConnectStatus(2)
)
