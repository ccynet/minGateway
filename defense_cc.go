package main

import (
	"minGateway/util/fixedqueue"
	"minGateway/util/log"
	"minGateway/util/tcache"
	"time"
)

//检查时间，毫秒
var defenseCCCheckTime int

//检查时间内的次数
var defenseCCCheckCount int

//加入黑名单时间，秒
var defenseCCBlackTime int

//IP记录
var defenseCCIPsRecord *tcache.TCache

//IP黑名单
var defenseCCIPsBlacklist *tcache.TCache

//配置中的黑白名单
var defenseCCIPWhiteListConf map[string]struct{}
var defenseCCIPBlackListConf map[string]struct{}

//判断是否存在于黑名单中
func existInCCBlacklist(ip string) bool {
	_, exist := defenseCCIPsBlacklist.Get(ip)
	return exist
}

//检查防御，如果返回true则加入黑名单
func defenseCCBlockCheck(ip string) bool {

	//配置中的白名单
	if _, exist := defenseCCIPWhiteListConf[ip]; exist {
		return false
	}
	//配置中的黑名单
	if _, exist := defenseCCIPBlackListConf[ip]; exist {
		return true
	}

	if existInCCBlacklist(ip) {
		return true
	}

	isBlock := false
	lt, ok := defenseCCIPsRecord.Get(ip)
	if !ok {
		record := fixedqueue.NewFixedQueue(defenseCCCheckCount)
		record.Push(makeTimestamp())
		defenseCCIPsRecord.Set(ip, record)
		return false
	} else {
		record := lt.(*fixedqueue.FixedQueue)
		if record.Len() == defenseCCCheckCount { //如果记录的长度已达到定义数
			//取出最老的那个时间
			oldest, _ := record.Get()
			now := makeTimestamp()
			if now-oldest.(int64) < int64(defenseCCCheckTime) {
				//当前时间减去最老的时间小于定义的时间限定
				//加入黑名单
				log.Warn(ip, "被加入黑名单", float64(defenseCCBlackTime)/60, "分钟")
				defenseCCIPsBlacklist.Set(ip, true)
				isBlock = true
			} else {
				record.Push(now)
			}
		} else {
			record.Push(makeTimestamp())
		}
	}
	return isBlock
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
