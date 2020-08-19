package main

import (
	"fmt"
	"io/ioutil"
	"minGateway/util"
	"sync"
	"testing"
	"time"
)

func TestExceededLimitReq(t *testing.T) {
	quit := make(chan bool)

	succNum := 0
	failNum := 0
	maxConn := make(chan bool, 5000)

	mu := &sync.Mutex{}

	go func() {
		time.Sleep(11 * time.Second)
		quit <- true
		return
	}()

	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-quit:
			fmt.Println("end")
			goto END
		case <-ticker.C:
			for i := 0; i < 20; i++ {
				go reqPing(&succNum, &failNum, mu, maxConn)
				time.Sleep(time.Millisecond * 10)
			}
		}
	}

END:
	fmt.Println("succ:", succNum, "  fail:", failNum)

}

func reqPing(succ, fail *int, mu *sync.Mutex, conn chan bool) {
	conn <- true
	req, err := util.HttpGet("http://127.0.0.1/api/ping?cid=100")
	<-conn
	if err != nil {
		fmt.Println("error:", err)
	} else {
		defer req.Body.Close()

		req, errC := ioutil.ReadAll(util.LimitReader(req.Body, 1024*1024))
		if errC != nil {
			fmt.Println("error:", err)
		}
		if string(req) == "pong" {
			mu.Lock()
			*succ++
			fmt.Println("pong")
			mu.Unlock()
			return
		}
	}
	mu.Lock()
	*fail++
	fmt.Println("fail")
	mu.Unlock()
	return
}
