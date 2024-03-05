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
	Base     Base
	Github   Github
	DingTalk DingTalk
}

type Base struct {
	Proxy       string        `ini:"proxy" comment:"默认为空,支持http和socks,最好不要包含;和#"`
	EnableProxy bool          `ini:"proxy_enable" comment:"默认:off"`
	Timeout     time.Duration `ini:"timeout" comment:"默认:10 * time.Second"`
	Interval    time.Duration `ini:"interval" comment:"默认:180 * time.Second"`
}

type Github struct {
	MaxRecordNumPerAuthor int `ini:"max_record_num_per_author" comment:"对于单条CVE"`
	MaxAuthorNumPerCve    int `ini:"max_author_num_per_cve" comment:"对于单条CVE"`
}

type DingTalk struct {
	DingTalkWebHookAccessToken string `ini:"dingtalk_webhook_access_token" comment:"钉钉聊机器人webhook的token，注意不是整个链接"`
	DingTalkWebHookSecret      string `ini:"dingtalk_webhook_secret" comment:"安全设置必须设置为”加签“，生成的sign"`
	Enable                     bool   `json:"enable"`
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
		Github: Github{MaxAuthorNumPerCve: 2, MaxRecordNumPerAuthor: 2},
		DingTalk: DingTalk{
			DingTalkWebHookAccessToken: "",
			DingTalkWebHookSecret:      "",
			Enable:                     false,
		},
	}
)

func GetSingleton() *Config {
	// 通过 sync.Once 确保仅执行一次实例化操作
	once.Do(func() {
	})
	return instance
}

func init() {
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