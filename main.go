package main

import (
	"cveHunter/config"
	"cveHunter/logger"
	"cveHunter/monitor/github"
	"cveHunter/proxy"
	"sync"
	"time"
)

func main() {
	//初始化配置文件
	config.GetSingleton()

	//使用代理
	proxy.GetSingleton().Add(github.GetSingleton())

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
