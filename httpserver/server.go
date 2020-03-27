package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.forceup.in/Payment/backend/l4g"
	"net/http"
	"time"
)

// 对外Http server
type (
	HttpServer struct {
		serverMux *http.ServeMux
		server    *http.Server
	}
)

func NewHttpServer() *HttpServer {
	hs := &HttpServer{
		serverMux: http.NewServeMux(),
	}

	return hs
}

func ResponseData(w http.ResponseWriter, d interface{}) {
	w.Write(fmtResponse(d, false))
}

func ResponseDataByIndent(w http.ResponseWriter, d interface{}) {
	w.Write(fmtResponse(d, true))
}

func fmtResponse(i interface{}, indent bool) []byte {
	switch d := i.(type) {
	case []byte:
		return d
	case string:
		return []byte(d)
	}

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	if indent {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(i); err != nil {
		return []byte(fmt.Sprintf("encoder data err: %s", err.Error()))
	}

	return bf.Bytes()
}

func (hs *HttpServer) Start(endpoint string) {
	//hs.serverMux.HandleFunc("/", indexHandler)
	//hs.serverMux.HandleFunc("/default", defaultHandler)

	hs.server = &http.Server{
		Addr:           endpoint,
		Handler:        hs.serverMux,
		ReadTimeout:    time.Second * 20,
		WriteTimeout:   time.Second * 20,
		MaxHeaderBytes: 0,
	}

	go func() {
		err := hs.server.ListenAndServe()
		if err != nil {
			l4g.BuildL4g("rpc2-center", "rpc2-center").Fatal(err.Error())
		}
	}()
}

func (hs *HttpServer) RegisterHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	hs.serverMux.HandleFunc(pattern, handler)
	return
}

func (hs *HttpServer) Stop() error {
	return hs.server.Close()
}
