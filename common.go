package main

import (
	"net/http"
	"strings"
)

var ipForwardeds []string

// 如果消息是通过前端代理服务器转发或者cdn转发，则需要从消息头中获取IP地址（注意确保IP的真实性），
// 如果消息直接来自于用户客户端，则使用req.RemoteAddr获取
func getIpAddr(req *http.Request) []string {
	if ipForwardeds == nil {
		return []string{strings.Split(req.RemoteAddr, ":")[0]}
	} else {
		for _, v := range ipForwardeds {
			if addr, ok := req.Header[v]; ok && len(addr) > 0 {
				return addr
			}
		}
		return []string{strings.Split(req.RemoteAddr, ":")[0]}
	}
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
