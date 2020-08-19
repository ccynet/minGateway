package main

import (
	"crypto/tls"
	"fmt"
	"github.com/g4zhuj/hashring"
	"golang.org/x/net/netutil"
	"math/rand"
	"minGateway/config"
	"minGateway/status"
	"minGateway/util/log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var (
	limitMaxConn   int //最大连接数
	readTimeout    int //读超时
	writeTimeout   int //写超时
	idleTimeout    int //闲置超时
	maxHeaderBytes int //最大头字节
)

type ObtainMode int //多转发目标时的选择模式

const (
	SelectModeRandom ObtainMode = 1 //随机选择
	SelectModePoll   ObtainMode = 2 //轮询选择
	SelectModeHash   ObtainMode = 3 //哈希选择
)

type HostInfoInterface interface {
	GetTarget(req *http.Request) string
}

type HostInfo struct {
	Target          string             //转发目标域名
	MultiTarget     []string           //有多转发目标的域名集合
	IsMultiTarget   bool               //是否有多转发目标
	MultiTargetMode ObtainMode         //多转发目标选择模式
	PoolModeIndex   int                //轮询模式索引
	hashRing        *hashring.HashRing //一致性哈希
}

func (hostInfo *HostInfo) GetTarget(req *http.Request) string {
	var route string
	if hostInfo.IsMultiTarget {
		if hostInfo.MultiTargetMode == SelectModeRandom { //随机模式
			route = hostInfo.MultiTarget[rand.Int()%len(hostInfo.MultiTarget)]
		} else if hostInfo.MultiTargetMode == SelectModePoll { //轮询模式
			route = hostInfo.MultiTarget[hostInfo.PoolModeIndex]
			hostInfo.PoolModeIndex++
			hostInfo.PoolModeIndex = hostInfo.PoolModeIndex % len(hostInfo.MultiTarget)
		} else if hostInfo.MultiTargetMode == SelectModeHash { //哈希模式
			ips := getIpAddr(req)
			route = hostInfo.hashRing.GetNode(ips[0])
		} else { //未配置或配置错误使用随机模式
			route = hostInfo.MultiTarget[rand.Int()%len(hostInfo.MultiTarget)]
		}
	} else {
		route = hostInfo.Target
	}
	return route
}

var HostList map[string]HostInfo

//通配符地址 wildcard character
type HostInfoWc struct {
	HostInfo
	//关键字的位置 0：前面 1：后面
	//关键字书写 如：wxpay. 那么KeyPos为0
	KeyPos int
}

var HostListWc map[string]HostInfoWc

//缺省转发，如果配置文件上定义了缺省转发，那么有消息进入时没找到已定义的转发地址就转发到缺省定义上
var DefaultTarget *HostInfo

var CertificateSet []tls.Certificate

type Proxy struct{}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("-> %s  ", r.Host)

	var ip string
	var ipsSize int
	if config.Get().CcDefense.Enable || config.Get().LimitReq.Enable {
		ips := getIpAddr(r)
		ipsSize = len(ips)
		if ipsSize > 0 {
			ip = ips[0]
		}
	}

	//如果配置开启CC防御
	if config.Get().CcDefense.Enable {
		//看是否在黑名单内，如果在黑名单内，直接返回
		if ipsSize > 0 {
			//检查是否符合阻挡规则
			if defenseCCBlockCheck(ip) {
				_ = r.Body.Close()
				return
			}
		}
	}

	//如果开启了请求数限制
	if config.Get().LimitReq.Enable {
		if ipsSize > 0 {
			// 是否超过限制请求数
			if ExceededLimitReq(ip, r) {
				return
			}
		}
	}

	in := time.Now()

	//设置状态：连接数，用于后面获取连接数
	status.Instance().AddReqCount()
	defer status.Instance().SubReqCount()

	//根据配置选择转发到哪
	var route string //转发的目标
	var existRoute = false
	if len(r.Host) == 0 {
		if DefaultTarget != nil {
			route = DefaultTarget.GetTarget(r)
			existRoute = true
		}
	} else if hostInfo, ok := HostList[r.Host]; ok {
		route = hostInfo.GetTarget(r)
		existRoute = true
	} else if len(HostListWc) > 0 {
		//轮询通配符集合，查找有没有符合的域名
		for likeHost, hostInfo := range HostListWc {
			if (hostInfo.KeyPos == 0 && strings.HasPrefix(r.Host, likeHost)) ||
				(hostInfo.KeyPos == 1 && strings.HasSuffix(r.Host, likeHost)) {
				route = hostInfo.GetTarget(r)
				existRoute = true
				break
			}
		}
	}
	if !existRoute {
		if DefaultTarget != nil {
			route = DefaultTarget.GetTarget(r)
			existRoute = true
		} else {
			log.Warnf("未配置的代理, %s", r.Host)
			return
		}
	}

	log.Debugf("-> %s", route)

	//找到转发目标，继续
	if existRoute {
		target, err := url.Parse(route)
		if err != nil {
			log.Error("url.Parse失败")
			return
		}

		proxy := newHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	}

	log.Debug("耗时：", time.Now().Sub(in).Seconds(), "秒")
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func newHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header["X-Real-Ip"] = getIpAddr(req)
		log.Debug("X-Real-Ip=", req.Header["X-Real-Ip"])
		//for k, v := range req.Header {
		//	log.Debug(k, v)
		//}
	}
	return &httputil.ReverseProxy{Director: director}
}

type GateServer struct{}

func (s *GateServer) proxy80() *http.Server {
	ln, err := net.Listen("tcp", ":80")
	if err != nil {
		panic(err)
	}

	if limitMaxConn > 0 {
		//限流
		ln = netutil.LimitListener(ln, limitMaxConn)
	}

	p := &Proxy{}
	srv := &http.Server{Addr: ":80", Handler: p}
	if readTimeout > 0 {
		srv.ReadTimeout = time.Duration(readTimeout) * time.Second
	}
	if writeTimeout > 0 {
		srv.WriteTimeout = time.Duration(writeTimeout) * time.Second
	}
	if idleTimeout > 0 {
		srv.IdleTimeout = time.Duration(idleTimeout) * time.Second
	}
	if maxHeaderBytes > 0 {
		srv.MaxHeaderBytes = maxHeaderBytes
	}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()
	fmt.Println("网关监听端口:80")

	return srv
}

func makeTlsConfig() *tls.Config {
	config1 := &tls.Config{Certificates: CertificateSet}
	config1.BuildNameToCertificate() //BuildNameToCertificate()使之能嗅探域名，如果没找到信息则使用数组[0]
	fmt.Println("\nSSL Set:", config1.NameToCertificate)
	return config1
}

func (s *GateServer) proxy443(tlsConfig *tls.Config) *http.Server {
	ln, err := tls.Listen("tcp", ":443", tlsConfig)
	if err != nil {
		panic(err)
	}

	if limitMaxConn > 0 {
		//限流
		ln = netutil.LimitListener(ln, limitMaxConn)
	}

	p := &Proxy{}
	srv := &http.Server{Addr: ":443", Handler: p}
	if readTimeout > 0 {
		srv.ReadTimeout = time.Duration(readTimeout) * time.Second
	}
	if writeTimeout > 0 {
		srv.WriteTimeout = time.Duration(writeTimeout) * time.Second
	}
	if idleTimeout > 0 {
		srv.IdleTimeout = time.Duration(idleTimeout) * time.Second
	}
	if maxHeaderBytes > 0 {
		srv.MaxHeaderBytes = maxHeaderBytes
	}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()
	fmt.Println("网关监听端口:443")

	return srv
}

func (s *GateServer) run() []*http.Server {

	ss := make([]*http.Server, 0)

	p80 := s.proxy80()
	ss = append(ss, p80)

	if len(CertificateSet) > 0 {

		tlsConfig := makeTlsConfig()

		// 是否开启SessionTicket，TLS1.3中即是否开启PSK
		// TLS1.3中SessionTicket报文也是加密的，我通过抓包无法看到 New session ticket 的报文，这里可能有问题，SessionTicket可能没起效
		if !config.Get().SslBase.SessionTicket {
			tlsConfig.SessionTicketsDisabled = true //禁止
		} else {
			sessiontickets := &SessionTicketService{}
			err := sessiontickets.Run()
			if err != nil {
				log.Error("SessionTicketService error", err)
			} else {
				sessiontickets.Register(tlsConfig)
			}
		}

		p443 := s.proxy443(tlsConfig)
		ss = append(ss, p443)
	}

	return ss
}
