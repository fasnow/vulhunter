package config

import (
	"cveHunter/logger"
	proxy2 "cveHunter/proxy"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	Base          Base
	Github        Github
	AVD           AVD
	DingTalk      DingTalk
	LarkAssistant LarkAssistant
	LarkBot       LarkBot
}

type Base struct {
	EnableProxy bool          `ini:"proxy_enable" comment:"默认:off"`
	Proxy       string        `ini:"proxy" comment:"默认为空,支持http和socks,最好不要包含;和#"`
	Timeout     time.Duration `ini:"timeout" comment:"默认:10 * time.Second"`
	Interval    time.Duration `ini:"interval" comment:"默认:180 * time.Second"`
}

type Github struct {
	Enable                bool   `ini:"enable"`
	GithubToken           string `ini:"github_token" comment:"搜索内容不包含代码: 未认证访问速率最高为10次/min, 认证后最高为30次/min,搜索内容包含代码: 必须认证,访问速率最高为10次/min"`
	MaxRecordNumPerAuthor int    `ini:"max_record_num_per_author" comment:"对于单条CVE"`
	MaxAuthorNumPerCve    int    `ini:"max_author_num_per_cve" comment:"对于单条CVE"`
}

type AVD struct {
	Enable bool `ini:"enable" comment:"阿里云漏洞库"`
}

type DingTalk struct {
	Enable             bool   `ini:"enable"`
	WebHookAccessToken string `ini:"webhook_access_token" comment:"钉钉群组机器人webhook的token，注意不是整个链接"`
	WebHookSecret      string `ini:"webhook_secret" comment:"安全设置必须设置为”加签“，生成的sign"`
}

type LarkAssistant struct {
	Enable             bool   `ini:"enable"`
	WebHookAccessToken string `ini:"webhook_access_token" comment:"飞书机器人助手webhook的token，注意不是整个链接"`
}

type LarkBot struct {
	Enable             bool   `ini:"enable"`
	WebHookAccessToken string `ini:"webhook_access_token" comment:"飞书群组机器人webhook的token，注意不是整个链接"`
	WebHookSecret      string `ini:"webhook_secret" comment:"安全设置必须设置为”签名校验“，生成的sign"`
}

var (
	instance      *Config
	once          sync.Once
	defaultConfig = &Config{
		Base: Base{
			Proxy:       "http://127.0.0.1:8080",
			EnableProxy: false,
			Timeout:     10 * time.Second,
			Interval:    180 * time.Second,
		},
		Github: Github{MaxAuthorNumPerCve: 2, MaxRecordNumPerAuthor: 2, Enable: true},
		AVD:    AVD{Enable: true},
	}
)

func GetSingleton() *Config {
	// 通过 sync.Once 确保仅执行一次实例化操作
	once.Do(func() {
	})
	return instance
}

func init() {
	ini.PrettyFormat = false
	options := ini.LoadOptions{
		SkipUnrecognizableLines:  true, //跳过无法识别的行
		SpaceBeforeInlineComment: true,
	}
	if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
		logger.Info("config file not found, generating default config file...")
		cfg := ini.Empty(options)
		err := ini.ReflectFrom(cfg, defaultConfig)
		if err != nil {
			logger.Info("can't reflect default config struct: " + err.Error())
			os.Exit(0)
		}
		err = cfg.SaveTo("config.ini")
		if err != nil {
			logger.Info("can't generate default config file: " + err.Error())
			os.Exit(0)
		}
		abs, err := filepath.Abs(filepath.Join("config.ini"))
		if err != nil {
			logger.Info("generated default config file successfully, located at config.ini")
			os.Exit(0)
		}
		logger.Info("generate default config file successfully, locate at " + abs + ", run with default config")
		instance = defaultConfig
		return
	}
	instance = defaultConfig

	//默认超时设置
	proxy2.GetSingleton().SetTimeout(instance.Base.Timeout)

	cfg, err := ini.LoadSources(
		options,
		"config.ini",
	)
	if err != nil {
		logger.Info("cat't open config file:" + err.Error())
		return
	}

	err = cfg.MapTo(defaultConfig)
	if err != nil {
		logger.Info("can't mapTo config file:" + err.Error())
		return
	}

	//代理
	if instance.Base.EnableProxy {
		if err = proxy2.GetSingleton().SetProxy(instance.Base.Proxy); err != nil {
			logger.Info("set proxy error: " + err.Error())
		}
	}

	//超时
	proxy2.GetSingleton().SetTimeout(time.Duration(instance.Base.Timeout) * time.Second)

}
