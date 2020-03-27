package common

import (
	"strings"
)

type (
	// 服务信息
	Service struct {
		Version string `json:"version"`
		Name    string `json:"name"`
		Tag     string `json:"tag"`
	}

	// 服务中心
	ConfigCenter struct {
		Service
		HttpPort  string   `json:"http_port"`
		RpcPort   string   `json:"rpc_port"`
		KeepAlive int      `json:"keep_alive"`
		Env       []string `json:"env"`
	}

	// 服务节点
	ConfigNode struct {
		Service
		RpcAddr string   `json:"rpc_addr"`
		Env     []string `json:"env"`
	}
)

// 获取服务唯一key(version.name)
func (s Service) GetKey() string {
	return strings.ToLower(s.Version + "." + s.Name)
}

// 获取服务唯一示例名称(version.name.tag)
func (s Service) GetInstance() string {
	return strings.ToLower(s.Version + "." + s.Name + "." + s.Tag)
}
