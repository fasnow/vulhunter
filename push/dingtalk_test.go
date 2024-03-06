package push

import (
	"bytes"
	"cveHunter/db"
	proxy2 "cveHunter/proxy"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestGetSingleton(t *testing.T) {
	c := GetDingTalkSingleton()
	proxy2.GetSingleton().Add(c)
	proxy2.GetSingleton().SetProxy("http://127.0.0.1:8080")
	c.Push([]db.GithubCVE{
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
	}...)
}

func TestGetSingleton1(t *testing.T) {
	cves := []db.GithubCVE{
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
	}

	var items []string
	for _, cve := range cves {
		items = append(items,
			fmt.Sprintf("**漏洞编号**:%s  \n**地址**:%s  \n**描述**:  \n  %s  \n", cve.Name, cve.HtmlUrl, cve.Description))
	}
	// 解析代理地址
	proxyURL, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		return
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	tt, _ := json.Marshal(map[string]any{
		//消息类型
		"msgtype": "markdown",

		//被@的群成员信息
		"at": map[string]any{
			"isAtAll":   "false",
			"atUserIds": []string{},
			"atMobiles": []string{},
		},

		//链接消息
		"link": map[string]any{
			"messageUrl": "1",
			"picUrl":     "1",
			"text":       "1",
			"title":      "1",
		},

		//markdown消息
		"markdown": map[string]any{
			"text":  strings.Join(items, "  \n------  \n"),
			"title": "Github漏洞推送",
		},

		//feedCard消息
		"feedCard": map[string]any{
			"links": map[string]any{
				"picURL":     "1",
				"messageURL": "1",
				"title":      "1",
			},
		},

		//文本消息
		"text": map[string]any{
			"content": "123",
		},

		//actionCard消息
		"actionCard": map[string]any{
			"hideAvatar":     "1",
			"btnOrientation": "1",
			"singleTitle":    "1",
			"btns": []any{map[string]any{
				"actionURL": "1",
				"title":     "1",
			}},
			"text":      "1",
			"singleURL": "1",
			"title":     "1",
		},
	})
	request, err := http.NewRequest("POST", "https://www.feishu.cn/flow/api/trigger-webhook/0636c51b995edd8f1516978bca411fae", bytes.NewReader(tt))
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")
	_, _ = httpClient.Do(request)

	return
}
