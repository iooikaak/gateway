package serve

import (
	"context"
	frame "github.com/iooikaak/frame/core"
	log "github.com/iooikaak/frame/log"
	gcfg "github.com/iooikaak/gateway/config"
	"github.com/iooikaak/gateway/session"
	"net"
	"net/http"
	"os"
)

var root = "/services"

// Gateway 网关服务
type Gateway struct {
	cfg        gcfg.Config
	listenAddr string
	azs        map[frame.Authorization]session.Authorization
	ac         session.Authorization
	rt         *Router
	discover   *frame.Discover
	Ctx        context.Context
}

// NewGateway 网关
func NewGateway(cfg gcfg.Config, ctx context.Context) *Gateway {
	gw := new(Gateway)
	gw.cfg = cfg
	gw.Ctx = ctx
	gw.azs = make(map[frame.Authorization]session.Authorization)
	for _, v := range cfg.Authorization {
		switch frame.Authorization(v) {
		case frame.App:
			gw.azs[frame.App] = session.NewApp(&cfg)
		case frame.Internal:
			gw.azs[frame.Internal] = session.NewInternal(&cfg)
		case frame.Openapi:
			gw.azs[frame.Openapi] = session.NewOpenapi(&cfg)
		default:
			log.Fatalf("Gateway not support authorization:%s", v)
			os.Exit(-1)
		}
	}
	//监听端口为内网ip和配置文件中的端口号
	gw.listenAddr = frame.Netip() + ":" + gw.cfg.Port
	/**
	* Start()持续开启服务发现
	* 自己注册到etcd里面并且通过Watch()持续监听
	* 新实例启动自己会注册到etcd里面,Watch()方法监听后会填加到内存map的键值对里面
	* 老实例优雅退出后,Watch()方法监听后会在内存map里面删除这个键值对
	 */
	gw.discover = frame.Instance().Start("Gateway", gw.cfg.Base, []string{root}, gw.listenAddr)
	//初始化serve.Router对象
	gw.rt = NewRouter(gw.DefaultHandler, HandleNotFound)
	gw.rt.Add("/ping", HandlePing)
	gw.rt.Add("/statistics", HandleStatics)
	gw.rt.Add("/metrics", PrometheusHandler)
	log.Debugf("gateway init with cfg=%v", gw.cfg)
	//返回gateway对象
	return gw
}

// ListenAndServe listen and serve
func (gw *Gateway) ListenAndServe() {
	l, err := net.Listen("tcp", gw.listenAddr)
	if err != nil {
		panic(err.Error())
	}
	http.Serve(l, gw)
}

// Stop Stop
func (gw *Gateway) Stop(s os.Signal) bool {
	log.Infof("gateway graceful exit.")
	return true
}

// ServeHTTP implement http ServeHTTP
// 服务入口，转发到来的所有请求到具体服务
func (gw *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gw.rt.Hanlder(r.URL.Path)(w, r)
}
