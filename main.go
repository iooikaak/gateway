package main

import (
	"context"
	frame "github.com/iooikaak/frame/core"
	"github.com/iooikaak/gateway/config"
	"github.com/iooikaak/gateway/serve"
	_ "net/http/pprof"
	"syscall"
)

func main() {
	//远程读取apollo配置文件
	//初始化apollo
	//if err := config.ApolloInit(); err != nil {
	//	panic(err)
	//}
	config.Init()
	//创建gateway对象
	s := serve.NewGateway(*config.Conf, context.Background())
	//初始化信号处理对象
	sh := frame.NewSignalHandler()
	//注册信号
	sh.Register(syscall.SIGTERM, s)
	sh.Register(syscall.SIGQUIT, s)
	sh.Register(syscall.SIGINT, s)
	//持续监听异常信号，如果有则退出程序
	sh.Start()
	s.ListenAndServe()
}
