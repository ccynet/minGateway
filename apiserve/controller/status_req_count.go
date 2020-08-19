package controller

import (
	"fmt"
	"minGateway/apiserve/comm"
	"minGateway/apiserve/resultcode"
	"minGateway/status"
	"net/http"
)

func StatusReqCount(w http.ResponseWriter, r *http.Request) {
	err := comm.Report(w, resultcode.SUCCESS, "success", comm.H{"count": status.Instance().GetReqCount()})
	if err != nil {
		fmt.Println(err)
	}
}
