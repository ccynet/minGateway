package main

import (
	"fmt"
	"github.com/g4zhuj/hashring"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	llog "log"
	"minGateway/util"
	"sync"
	"testing"
	"time"
)

func TestHashRing(t *testing.T) {
	ips := []string{}
	for i := 0; i < 10000; i++ {
		ips = append(ips, fmt.Sprint("12.13.456.", i))

	}

	nodeWeight := make(map[string]int)
	nodeWeight["node1"] = 1
	nodeWeight["node2"] = 1
	nodeWeight["node3"] = 2
	vitualSpots := 100
	hash := hashring.NewHashRing(vitualSpots)

	//add nodes
	hash.AddNodes(nodeWeight)

	node1Count := 0
	node2Count := 0
	node3Count := 0
	//get key's node
	for _, ip := range ips {
		node := hash.GetNode(ip)
		if node == "node1" {
			node1Count++
		}
		if node == "node2" {
			node2Count++
		}
		if node == "node3" {
			node3Count++
		}
	}

	fmt.Println("node1=", node1Count, " node2=", node2Count, " node3=", node3Count)

	fmt.Println(hash.GetNode(""))

}

func TestTimeD(t *testing.T) {
	in := time.Now()
	time.Sleep(100 * time.Millisecond)
	t.Log("耗时：", time.Now().Sub(in).Seconds(), "秒")
}

func TestBytes(t *testing.T) {
	b := 1 << 20
	t.Log(b, b/1024, "k")
	b = 1 << 17
	t.Log(b, b/1024, "k")
}

func TestPing(t *testing.T) {

	quit := make(chan bool)

	succNum := 0
	failNum := 0
	maxConn := make(chan bool, 5000)

	mu := &sync.Mutex{}

	go func() {
		time.Sleep(1 * time.Minute)
		quit <- true
		return
	}()

	ticker := time.NewTicker(20 * time.Millisecond)

	for {
		select {
		case <-quit:
			fmt.Println("end")
			goto END
		case <-ticker.C:
			go sendPing(&succNum, &failNum, mu, maxConn)
		}
	}

END:
	fmt.Println("succ:", succNum, "  fail:", failNum)

}

func sendPing(succ, fali *int, mu *sync.Mutex, conn chan bool) {
	//req, err := util.HttpGet("http://alogin.obtc.com/zgd/api_sys/ping")
	conn <- true
	req, err := util.HttpGet("http://139.199.75.120:8080/zgd/api_sys/ping")
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
			mu.Unlock()
			return
		}
	}
	mu.Lock()
	*fali++
	mu.Unlock()
	return
}

var printTagCh = make(chan int, 1000)

func BenchmarkLogrusChan(b *testing.B) {
	type args struct {
		level int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test logrus.level 1",
			args: args{level: 1},
		},
		//{
		//	name: "test logrus.level 2",
		//	args: args{level: 2},
		//},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			l := logrus.New()

			//l.SetNoLock()

			if tt.args.level == 1 {
				l.SetLevel(logrus.DebugLevel)
			} else if tt.args.level == 2 {
				l.SetLevel(logrus.InfoLevel)
			}

			for j := 0; j < b.N; j++ {
				printTagCh <- j
				go func() {
					l.Debug(<-printTagCh)
				}()
			}
			//b.Log(l)
		})
	}
}

func BenchmarkLogrus(b *testing.B) {
	type args struct {
		level int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test logrus.level 1",
			args: args{level: 1},
		},
		{
			name: "test logrus.level 2",
			args: args{level: 2},
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			l := logrus.New()

			if tt.args.level == 1 {
				l.SetLevel(logrus.DebugLevel)
			} else if tt.args.level == 2 {
				l.SetLevel(logrus.InfoLevel)
			}

			for j := 0; j < b.N; j++ {
				l.Debug("info info info info info info")
			}
			//b.Log(l)
		})
	}
}

func BenchmarkSysLog(b *testing.B) {
	type args struct {
		level int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test logrus.level 1",
			args: args{level: 1},
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {

			for j := 0; j < b.N; j++ {
				llog.Println("info info info info info info")
				fmt.Println("info info info info info info")
			}
			//b.Log(l)
		})
	}
}

func BenchmarkFmtLog(b *testing.B) {
	type args struct {
		level int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test logrus.level 1",
			args: args{level: 1},
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {

			for j := 0; j < b.N; j++ {
				fmt.Println("info info info info info info")
			}
			//b.Log(l)
		})
	}
}
