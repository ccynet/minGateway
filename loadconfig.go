package main

import (
	"crypto/tls"
	"fmt"
	"github.com/g4zhuj/hashring"
	"io/ioutil"
	"minGateway/config"
	"minGateway/util"
	"minGateway/util/tcache"
	"strings"
	"time"
)

func loadConfig() error {
	coreConf := config.Get().CoreConf
	limitMaxConn = coreConf.LimitMaxConn
	readTimeout = coreConf.ReadTimeout
	writeTimeout = coreConf.WriteTimeout
	idleTimeout = coreConf.IdleTimeout
	maxHeaderBytes = coreConf.MaxHeaderBytes
	if coreConf.IpForwardeds != nil && len(coreConf.IpForwardeds) > 0 {
		ipForwardeds = coreConf.IpForwardeds
	}

	//如果定义了activeDefense即开启防御
	if config.Get().CcDefense.Enable {
		//监测的时间限定，毫秒
		defenseCCCheckTime = config.Get().CcDefense.TimeDuration

		//在限定时间内最大次数
		defenseCCCheckCount = config.Get().CcDefense.Count

		//放入黑名单中的时间，秒
		defenseCCBlackTime = config.Get().CcDefense.BlackTime

		//ip记录保存10小时
		defenseCCIPsRecord = tcache.NewTimeCache(10 * time.Hour)

		//ip黑名单保存时间
		defenseCCIPsBlacklist = tcache.NewTimeCache(time.Duration(defenseCCBlackTime) * time.Second)

		//配置黑白名单
		defenseCCIPWhiteListConf = make(map[string]struct{})
		defenseCCIPBlackListConf = make(map[string]struct{})
		for _, ip := range config.Get().CcDefense.WhiteList {
			defenseCCIPWhiteListConf[ip] = struct{}{}
		}
		for _, ip := range config.Get().CcDefense.BlackList {
			defenseCCIPBlackListConf[ip] = struct{}{}
		}
	}

	//请求限制配置
	if config.Get().LimitReq.Enable {
		limitReqMap = tcache.NewTimeCache(time.Millisecond * time.Duration(config.Get().LimitReq.TimeDuration))
	}

	httpConfig := config.Get().HttpProxy
	for _, v := range httpConfig {

		info := getHostInfo(v)

		if v.Host == "default" {
			//如果定义了default, 遇到未知host走这里
			DefaultTarget = &info
		} else if strings.HasSuffix(v.Host, ".") {
			//关键字在"."的前面
			wc := HostInfoWc{KeyPos: 0}
			wc.IsMultiTarget = info.IsMultiTarget
			wc.MultiTarget = info.MultiTarget
			wc.MultiTargetMode = info.MultiTargetMode
			wc.Target = info.Target
			if info.hashRing != nil {
				wc.hashRing = info.hashRing
			}
			HostListWc[v.Host] = wc
		} else if strings.HasPrefix(v.Host, ".") {
			//关键字在"."的后面
			wc := HostInfoWc{KeyPos: 1}
			wc.IsMultiTarget = info.IsMultiTarget
			wc.MultiTarget = info.MultiTarget
			wc.MultiTargetMode = info.MultiTargetMode
			wc.Target = info.Target
			if info.hashRing != nil {
				wc.hashRing = info.hashRing
			}
			HostListWc[v.Host] = wc
		} else {
			HostList[v.Host] = info
			if strings.HasPrefix(v.Host, "www.") {
				if strings.Count(v.Host, ".") == 2 {
					//一级域名，考虑没有带"www"的情况
					HostList[strings.TrimLeft(v.Host, "www.")] = HostList[v.Host]
				}
			} else if strings.Count(v.Host, ".") == 1 {
				//排除首位和末位的"."，"."的数量只有一个说明是没有带"www"的一级域名
				HostList["www."+v.Host] = HostList[v.Host]
			}
		}
	}
	fmt.Println()
	fmt.Println("监听的反向代理域名：")
	for k, v := range HostList {
		if v.IsMultiTarget {
			fmt.Println(k, "->", v.MultiTarget, " mode:", v.MultiTargetMode)
		} else {
			fmt.Println(k, "->", v.Target)
		}
	}
	for k, v := range HostListWc {
		if v.IsMultiTarget {
			fmt.Println(k, "->", v.MultiTarget, " mode:", v.MultiTargetMode)
		} else {
			fmt.Println(k, "->", v.Target)
		}
	}
	if DefaultTarget != nil {
		if DefaultTarget.IsMultiTarget {
			fmt.Println("default", "->", DefaultTarget.MultiTarget, " mode:", DefaultTarget.MultiTargetMode)
		} else {
			fmt.Println("default", "->", DefaultTarget.Target)
		}
	}

	fmt.Println()
	fmt.Println("SSL证书：")
	sslConfig := config.Get().SslCert
	for k, v := range sslConfig {
		certFile := util.GetAbsolutePath("/bin/cert/%s", v.SslCertificate)
		keyFile := util.GetAbsolutePath("/bin/cert/%s", v.SslCertificateKey)
		crt, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println(k, ":")
		fmt.Println(certFile)
		fmt.Println(keyFile)

		//OCSP Stapling
		//在crt.OCSPStaple上附值则表示开启OCSP封套
		if v.OcspStapling {
			//是读取本地缓存文件还是在线获取
			if v.OcspStaplingLocal {
				ocspFile := util.GetAbsolutePath("/bin/cert/%s", v.OcspStaplingFile)
				OCSPBuf, err := ioutil.ReadFile(ocspFile)
				if err == nil {
					crt.OCSPStaple = OCSPBuf
					fmt.Println(k, ", local load, OCSP Stapling OK")
				} else {
					fmt.Println(err)
				}
			} else {
				OCSPBuf, _, _, err := GetOCSPForCert(crt.Certificate)
				if err == nil {
					crt.OCSPStaple = OCSPBuf
					fmt.Println(k, ", online load, OCSP Stapling OK")
				}
			}
		}

		CertificateSet = append(CertificateSet, crt)

	}

	fmt.Println()

	return nil
}

func getHostInfo(proxyInfo config.ProxyInfo) HostInfo {
	var hostInfo HostInfo
	for i := 0; i < len(proxyInfo.Target); i++ {
		proxyInfo.Target[i] = strings.ReplaceAll(proxyInfo.Target[i], " ", "")
	}
	if len(proxyInfo.Target) == 0 {
		panic(proxyInfo.Host + " : len(proxyInfo.Target) == 0")
	} else if len(proxyInfo.Target) == 1 {
		hostInfo = HostInfo{IsMultiTarget: false, Target: proxyInfo.Target[0]}
	} else {
		//定义了多个目标，使用分流
		targets := proxyInfo.Target
		hostInfo = HostInfo{IsMultiTarget: true, MultiTarget: targets, MultiTargetMode: ObtainMode(proxyInfo.ObtainMode)}
		if proxyInfo.ObtainMode == 3 { //哈希模式
			//把节点放到hashring中，同时设置权重
			hostInfo.hashRing = hashring.NewHashRing(100) //vitualSpots=100
			nodeWeight := make(map[string]int)
			for _, target := range targets {
				nodeWeight[target] = 1 //这里简化了，权重都设置为1
			}
			hostInfo.hashRing.AddNodes(nodeWeight)
		}
	}
	return hostInfo
}
