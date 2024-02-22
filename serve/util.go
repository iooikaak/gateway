package serve

import (
	"fmt"
	frame "github.com/iooikaak/frame/core"
	"github.com/iooikaak/frame/protocol"
	"github.com/nobugtodebug/go-objectid"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
)

var defaultMemory = int64(32 >> 22)

func resolveRequest(r *http.Request) (*protocol.Proto, error) {
	// 准备warhorse通用协议
	var task protocol.Proto
	businessID := r.Header.Get("bizid")
	if businessID == "" {
		businessID = objectid.New().String()
	}
	path := r.URL.Path
	// Tracer ID
	task.Bizid = businessID
	i := strings.LastIndexByte(path, '/')
	if i == -1 {
		return nil, fmt.Errorf("error path:%s", path)
	}
	// URI
	task.ServeURI = path[:i]
	// Serve Method
	task.ServeMethod = strings.ToLower(path[i+1:])
	// Inner ID
	task.RequestID = frame.GetInnerID()
	// HTTP Method
	n := protocol.RestfulMethod_value[strings.ToUpper(r.Method)]
	task.Method = protocol.RestfulMethod(n)
	task.RemoteAddr = r.RemoteAddr
	// 解析Header信息
	task.Header = make(map[string]string)
	for k := range r.Header {
		task.Header[k] = r.Header.Get(k) //jaeger记录此次追踪
	}
	// URL RAW Query data
	task.Form = make(map[string]string)
	q := r.URL.Query()
	for k := range q {
		task.Form[k] = q.Get(k)
	}
	body, _ := io.ReadAll(r.Body)
	task.Body = body
	// 解析Form，Body信息
	contentType := r.Header.Get(protocol.HeaderContentType)
	if strings.HasPrefix(contentType, protocol.MIMEApplicationJSON) {
		task.Format = protocol.RestfulFormat_JSON
		return &task, nil
	}
	if strings.HasPrefix(contentType, protocol.MIMETextXML) || strings.HasPrefix(contentType, protocol.MIMEApplicationXML) {
		task.Format = protocol.RestfulFormat_XML
		return &task, nil
	}
	if strings.HasPrefix(contentType, protocol.MIMEApplicationProtobuf) {
		task.Format = protocol.RestfulFormat_PROTOBUF
		return &task, nil
	}
	if strings.HasPrefix(contentType, protocol.MIMEMultipartForm) {
		if err := r.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	}
	if strings.HasPrefix(contentType, protocol.MIMEApplicationForm) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
	}
	if len(r.Form) > 0 {
		for k := range r.Form {
			task.Form[k] = r.Form.Get(k)
		}
		task.Format = protocol.RestfulFormat_RAWQUERY
		return &task, nil
	}
	task.Format = protocol.RestfulFormat_FORMATNULL
	return &task, nil
}

func getConnNetErrno(err error) (syscall.Errno, *net.OpError) {
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
				if errno, ok := sysErr.Err.(syscall.Errno); ok {
					return errno, opErr
				}
			}
		}
	}
	return 0, nil
}
