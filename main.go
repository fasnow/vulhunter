package main

import (
	"cveHunter/config"
	"cveHunter/logger"
	"cveHunter/monitor/github"
	"cveHunter/proxy"
	"cveHunter/push"
	"sync"
	"time"
)

func main() {
	//初始化配置文件
	config.GetSingleton()

	//使用代理的模块
	proxy.GetSingleton().Add(
		github.GetSingleton(),
		push.GetDingTalkSingleton(),
	)

	//设置代理
	err := proxy.GetSingleton().SetProxy("http://127.0.0.1:8080")
	if err != nil {
		logger.Info(err.Error())
	}

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	logger.Info("service is running...")
	go func() {
		defer waitGroup.Done()
		for {
			Run()
			logger.Info("waiting for next loop...")
			time.Sleep(config.GetSingleton().Base.Interval)
		}
	}()
	waitGroup.Wait()
}
