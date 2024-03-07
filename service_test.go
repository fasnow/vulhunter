package main

import (
	"cveHunter/config"
	"cveHunter/logger"
	"cveHunter/monitor"
	"cveHunter/proxy"
	"cveHunter/push"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	//初始化配置文件
	configSingleton := config.GetSingleton()

	pusherList := []string{}
	if configSingleton.DingTalk.Enable {
		pusherList = append(pusherList, "DingTalk")
	}
	if configSingleton.LarkAssistant.Enable {
		pusherList = append(pusherList, "LarkAssistant")
	}
	if configSingleton.LarkBot.Enable {
		pusherList = append(pusherList, "LarkBot")
	}
	if len(pusherList) == 0 {
		logger.Info("didn't enable any pusher,won't push any event")
	} else {
		logger.Info(fmt.Sprintf("enabled pusher [%s]", strings.Join(pusherList, ", ")))
	}

	//使用代理的模块
	proxy.GetSingleton().Add(
		monitor.GetGithubSingleton(),
		monitor.GetAVDMonitorSingleton(),
		push.GetDingTalkSingleton(),
		push.GetLarkAssistantLarkSingleton(),
		push.GetLarkSingleton(),
	)

	//设置代理
	err := proxy.GetSingleton().SetProxy("http://127.0.0.1:8080")
	if err != nil {
		logger.Info(err.Error())
	}

	waitGroup := &sync.WaitGroup{}

	logger.Info("service is running...")
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		if config.GetSingleton().Github.Enable {
			for {
				RunGithubMonitor()
				logger.Info("waiting for next loop...")
				time.Sleep(configSingleton.Base.Interval)
			}
		}
	}()
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		if config.GetSingleton().AVD.Enable {
			for {
				RunAVDMonitor()
				logger.Info("waiting for next loop...")
				time.Sleep(configSingleton.Base.Interval)
			}
		}
	}()
	waitGroup.Wait()
}
