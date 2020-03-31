package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// 对外输入输出
type (
	HttpUserResponse struct {
		Err    ErrCode     `json:"err" doc:"错误码"`
		ErrMsg string      `json:"errmsg,omitempty" doc:"错误信息"`
		Result interface{} `json:"result,omitempty" doc:"返回数据，需要自己解析"`
	}
)

// 内部RPC
type (
	Context struct {
	}

	Register struct {
		Service
		StartAt      string            `json:"start_at"`
		Meta         map[string]string `json:"meta"`
		Env          map[string]string `json:"env"`
		CallerList   []string          `json:"caller_list"`
		NotifierList []string          `json:"notifier_list"`
	}

	Method struct {
		Service
		Function string `json:"function"`
	}

	// IMPORTANT!!! do not directly set Value=..., use SetValue and GetValue
	UserRequest struct {
		Value string `json:"value,omitempty" doc:"请求数据，需要自己解析"`
	}

	// IMPORTANT!!! do not directly set Result=..., use SetResult and GetResult
	UserResponse struct {
		Err    ErrCode `json:"err" doc:"错误码"`
		ErrMsg string  `json:"errmsg,omitempty" doc:"错误信息"`
		Result string  `json:"result,omitempty" doc:"返回数据，需要自己解析"`
	}

	Request struct {
		Context Context     `json:"context"`
		Method  Method      `json:"method"`
		Data    UserRequest `json:"data"`
	}

	Response struct {
		Context Context      `json:"context"`
		Data    UserResponse `json:"data"`
	}
)

func (self *Response) Error() error {
	if self.Data.Err == ErrOk {
		return nil
	}

	var message string
	if self.Data.ErrMsg != "" {
		message = self.Data.ErrMsg
	} else {
		message = self.Data.Err.String()
	}
	return fmt.Errorf("err_code:%d, message:%s", self.Data.Err, message)
}

func (self *Response) SetResult(code ErrCode, err_msg string) {
	self.Data.Err = code
	self.Data.ErrMsg = err_msg
}

// 从path解析方法
func (method *Method) FromPath(path string) {
	path = strings.Trim(path, "/")
	paths := strings.Split(path, "/")
	for i := 0; i < len(paths); i++ {
		if i == 1 {
			method.Version = paths[i]
		} else if i == 2 {
			method.Name = paths[i]
		} else if i >= 3 {
			if method.Function != "" {
				method.Function += "."
			}
			method.Function += paths[i]
		}
	}
}

func (req *UserRequest) SetValue(d interface{}) error {
	var err error
	req.Value, err = toData(d)
	return err
}

func (req *UserRequest) GetValue(value interface{}) error {
	return fromData(req.Value, &value)
}

func (res *UserResponse) SetResult(d interface{}) error {
	var err error
	res.Result, err = toData(d)
	return err
}

func (res *UserResponse) GetResult(result interface{}) error {
	return fromData(res.Result, &result)
}

func toData(value interface{}) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func fromData(data string, value interface{}) error {
	if data == "" {
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &value)
}
