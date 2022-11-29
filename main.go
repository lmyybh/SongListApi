package main

import (
	"net/http"
	"songlist/managers"
	"songlist/routers"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("fun", "main")

func main() {
	//加载配置文件
	managers.Environment()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go managers.InitDatabase(&wg) // 初始化数据库
	go managers.InitRedis(&wg)    // 初始化 redis

	//等待初始化完成
	wg.Wait()

	routers.Index()

	// 开始服务
	port := strconv.Itoa(managers.CONFIG.Port)
	log.Info("Service started at port: " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Panic("Listen port failed", err)
	}
}
