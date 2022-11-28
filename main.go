package main

import (
	"net/http"
	"songlist/managers"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("fun", "main")

func ping(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Pong!"))
}

func main() {
	//加载配置文件
	managers.Environment()

	wg := sync.WaitGroup{}
	wg.Add(1)
	// 初始化数据库
	go managers.InitDatabase(&wg)
	//等待初始化完成
	wg.Wait()

	http.Handle("/ping", http.HandlerFunc(ping))

	// 开始服务
	port := strconv.Itoa(managers.CONFIG.Port)
	log.Info("Service started at port: " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Panic("Listen port failed", err)
	}
}
