package status

import (
	"sync"
)

var status *Status

type Status struct {
	// 连接数
	ReqCount   int
	ReqCountMu *sync.RWMutex
}

func Init() error {
	status = &Status{
		ReqCount:   0,
		ReqCountMu: new(sync.RWMutex),
	}
	return nil
}

func Instance() *Status {
	return status
}

func (sta *Status) AddReqCount() {
	sta.ReqCountMu.Lock()
	sta.ReqCount++
	sta.ReqCountMu.Unlock()
}

func (sta *Status) SubReqCount() {
	sta.ReqCountMu.Lock()
	sta.ReqCount--
	sta.ReqCountMu.Unlock()
}

func (sta *Status) GetReqCount() int {
	sta.ReqCountMu.RLock()
	defer sta.ReqCountMu.RUnlock()
	return sta.ReqCount
}
