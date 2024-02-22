package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/iooikaak/frame/stat/metric/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	frame "github.com/iooikaak/frame/core"

	log "github.com/iooikaak/frame/log"
	"github.com/iooikaak/frame/protocol"

	"go.uber.org/zap"
)

var (
	m = metrics.New("gateway")
)

const (
	warhorseFormatPattern = "%s - \"%s %s %d\" %s %s %s"
)

func PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	//promhttp.Handler().ServeHTTP(w,r)
	h := promhttp.InstrumentMetricHandler(m.RegInstance(), promhttp.HandlerFor(m.Gather(), promhttp.HandlerOpts{}))
	h.ServeHTTP(w, r)
}

// HandlePing LBS ping
func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("success"))
}

// HandleStatics 接口访问统计
func HandleStatics(w http.ResponseWriter, r *http.Request) {
	mt := frame.MeTables()
	if b, err := json.Marshal(mt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(b)
	}
}

// HandleNotFound http 404
func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	log.Warnf(warhorseFormatPattern, r.RemoteAddr, r.Method, r.URL.Path, r.ContentLength, r.Header.Get("Referer"), r.Header.Get("User-Agent"), "")
	http.Error(w, errNotFounHTTPMethod, http.StatusNotFound)
}

// DefaultHandler warhorse 默认的鉴权转发方式
func (gw *Gateway) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	var authType frame.Authorization
	//前端传入的urlPath匹配不到内存mdtables(map),报404
	if authType = frame.Instance().Authorization(r.URL.Path); string(authType) == "" {
		http.Error(w, "Illegal Intentions 非法意图! ", http.StatusNotFound)
		return
	}
	v, ok := gw.azs[authType]
	//urlPath的鉴权模式(app,internal,openapi)在内存中azs(map)支持的鉴权模式找不到
	if !ok {
		log.Warnf(warhorseFormatPattern, r.RemoteAddr, r.Method, r.URL.Path, r.ContentLength, r.Header.Get("Referer"), r.Header.Get("User-Agent"), authType)
		http.Error(w, "Illegal Intentions 非法意图!!", http.StatusForbidden)
		return
	}
	//获取鉴权模式(app,internal,openapi)并执行对应的鉴权方法
	if err := v.Do(r); err != nil {
		log.Warn(err.Error())
		http.Error(w, "Illegal Intentions 非法意图!!!", http.StatusUnauthorized)
		return
	}
	//组装protocol.Proto结构体并打印日志
	task, err := resolveRequest(r)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, errRequestInvalide, http.StatusBadRequest)
		return
	}
	go func() {
		log.Default().Zap.Info("req info", zap.String("Bizid", task.Bizid), zap.Int64("RequestID", task.RequestID),
			zap.String("ServeURI", task.ServeURI), zap.Any("Format", task.Format), zap.String("ServeMethod", task.ServeMethod),
			zap.Any("Method", task.Method), zap.Any("Header", task.Header), zap.Any("Form", task.Form),
			zap.ByteString("Body", task.Body), zap.ByteString("Err", task.Err), zap.String("RemoteAddr", task.RemoteAddr),
			zap.Any("TraceMap", task.TraceMap))
	}()
	//转发到具体服务并接收返回的protocol.Proto结构体
	resp, err := frame.DeliverTo(task)
	if err != nil {
		log.Warn(err.Error())
		if err == frame.ErrTimeout {
			http.Error(w, errGatewayTimeout, http.StatusGatewayTimeout)
			return
		}
		//connection refused 移除节点
		errno, opError := getConnNetErrno(err)
		url := r.URL.String()
		if errno == syscall.ECONNREFUSED {
			ns := strings.Split(url, "/")
			nsl := len(ns)
			interfaceURI := strings.Join(ns[:nsl-1], "/")
			fmt.Print(interfaceURI, opError)
			//todo 熔断 代替直接移除节点
			//gw.discover.RemoveNode(interfaceURI, opError.Addr.String())
		}
		/*
			多个服务中最新部署的下线，最新部署的服务的key 会覆盖，然后etcd会通知删除类似
			/services/v1/service/provider/name,
			/services/v1/service/method/provider/authorization/internal，
			服务列表只在gateway内存中存储。
			如果资源不存在，表示服务都已经下线可以清除
		*/
		gw.discover.RemoveMdTables(url, 1)
		log.Debugf("DeliverTo error Check MdTables %s", url)
		http.Error(w, errInternalError, http.StatusInternalServerError)
		return
	}
	//记录返回的protocol.Proto结构体
	go func() {
		log.Default().Zap.Info("resp info", zap.String("Bizid", resp.Bizid), zap.Int64("RequestID", resp.RequestID),
			zap.String("ServeURI", resp.ServeURI), zap.Any("Format", resp.Format), zap.String("ServeMethod", resp.ServeMethod),
			zap.Any("Method", resp.Method), zap.Any("Header", resp.Header), zap.Any("Form", resp.Form),
			zap.ByteString("Body", resp.Body), zap.ByteString("Err", resp.Err), zap.String("RemoteAddr", task.RemoteAddr),
			zap.Any("TraceMap", task.TraceMap))
	}()
	//响应结果的错误信息不为空
	if len(resp.GetErr()) > 0 {
		var errResponse protocol.Message
		if err := json.Unmarshal(resp.GetErr(), &errResponse); err != nil {
			http.Error(w, string(resp.Err), http.StatusInternalServerError)
			return
		}
		http.Error(w, errResponse.Errmsg, errResponse.Errcode)
		return
	}
	//正常返回，响应结果在请求体里面
	if len(resp.GetBody()) > 0 {
		for k, v := range resp.GetHeader() {
			w.Header().Set(k, v)
		}
		w.Write(resp.GetBody())
		return
	}
	//未知错误,返回的响应体为空并且返回没有错误信息
	http.Error(w, errInternalError, http.StatusInternalServerError)
	return
}
