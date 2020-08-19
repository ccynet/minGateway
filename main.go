package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"minGateway/apiserve"
	"minGateway/config"
	"minGateway/status"
	"minGateway/util"
	"minGateway/util/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	//加载配置
	err := loadConfig()
	if err != nil {
		panic(err)
	}

	//log设置
	logSetting()

	//运行服务
	srv := new(GateServer)
	ss := srv.run()

	//运行API服务
	api := apiserve.Run()
	ss = append(ss, api)

	//wait exit
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit
	fmt.Println("Start shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, s := range ss {
		if err := s.Shutdown(ctx); err != nil {
			fmt.Printf("Server Shutdown error, signal:%v error:%s\n", sig, err)
		}
	}

	fmt.Printf("Safe exit server, signal:%v\n", sig)
}

func init() {
	HostList = make(map[string]HostInfo)
	HostListWc = make(map[string]HostInfoWc)
	CertificateSet = make([]tls.Certificate, 0)

	// 状态信息初始化
	_ = status.Init()
}

func logSetting() {
	logConf := config.Get().LogConf

	// 设置日志输出级别
	// DebugLevel,InfoLevel,WarnLevel,ErrorLevel,FatalLevel,PanicLevel
	log.Init(logConf.LogLevel)

	// 是否写日志文件
	if logConf.WriteFile {

		logFile := "logging.log"
		if logConf.FileDir != "" {
			//FileDir不为空使用绝对路径
			logFile = logConf.FileDir + logFile
		} else {
			dir := util.GetAbsolutePath("/bin/logs")
			if !util.PathExists(dir) {
				_ = os.Mkdir(dir, 0666)
			}
			logFile = util.GetAbsolutePath("/bin/logs/%s", logFile)
		}

		log.WirteLog(logFile)
	}
}
