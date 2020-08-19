package main

import (
	"minGateway/config"
	"minGateway/util"
	"minGateway/util/tcache"
	"net/http"
)

var limitReqMap *tcache.TCache

func getReqCount(key string) int {
	hashkey := util.GetMD5(key)
	obj, ok := limitReqMap.Get(hashkey)
	if ok {
		count := obj.(int)
		limitReqMap.Set(hashkey, count+1)
		return count
	} else {
		limitReqMap.Set(hashkey, 1)
	}
	return 0
}

func ExceededLimitReq(ip string, r *http.Request) bool {
	var key string
	if config.Get().LimitReq.Mode == 0 {
		key = ip + r.Host + r.URL.Path
	} else {
		key = ip + r.RequestURI
	}
	count := getReqCount(key)
	if count >= config.Get().LimitReq.Count {
		return true
	} else {
		return false
	}
}
