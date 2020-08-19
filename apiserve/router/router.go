package router

import (
	"github.com/gorilla/mux"
	"minGateway/apiserve/controller"
)

func InitRouter(router *mux.Router) {
	if router == nil {
		panic("mux.Router is nil")
	}

	//获取状态
	router.HandleFunc("/api_status/reqcount", controller.StatusReqCount)
}
