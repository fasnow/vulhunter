package push

import (
	"bytes"
	"cveHunter/config"
	"cveHunter/db"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type LarkAssistant struct {
	webhookToken string
	HttpClient   *http.Client
}

var (
	larkAssistantPusher *LarkAssistant
	larkAssistantOnce   sync.Once
)

func GetLarkAssistantLarkSingleton() *LarkAssistant {
	// 通过 sync.Once 确保仅执行一次实例化操作
	larkAssistantOnce.Do(func() {
		larkAssistantPusher = &LarkAssistant{
			webhookToken: config.GetSingleton().LarkAssistant.WebHookAccessToken,
			HttpClient:   &http.Client{},
		}
	})
	return larkAssistantPusher
}

// Push 小于0表示推送失败
func (r *LarkAssistant) Push(cves ...db.GithubCVE) (int, error) {
	if r.webhookToken == "" {
		return -1, fmt.Errorf("no webhook_access_token")
	}
	var items []string
	for _, cve := range cves {
		msg := fmt.Sprintf("**漏洞编号**:%s  \n**地址**:%s  \n**描述**:", cve.Name, cve.HtmlUrl)
		if cve.Description != "" {
			msg = fmt.Sprintf("%s<font color='grey'>%s</font>", msg, cve.Description)
		}
		items = append(items, msg)
	}

	//https://www.feishu.cn/hc/zh-CN/articles/807992406756-webhook-%E8%A7%A6%E5%8F%91%E5%99%A8
	//https: //www.feishu.cn/hc/zh-CN/articles/236028437163-%E6%9C%BA%E5%99%A8%E4%BA%BA%E6%B6%88%E6%81%AF%E5%86%85%E5%AE%B9%E6%94%AF%E6%8C%81%E7%9A%84%E6%96%87%E6%9C%AC%E6%A0%B7%E5%BC%8F
	t, _ := json.Marshal(map[string]any{
		//消息类型
		"title": "Github漏洞推送",

		//markdown消息
		"content": strings.Join(items, "\n\n"),
	})
	request, err := http.NewRequest("POST", "https://www.feishu.cn/flow/api/trigger-webhook/"+r.webhookToken, bytes.NewReader(t))
	if err != nil {
		return -1, err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := r.HttpClient.Do(request)
	if err != nil {
		return -1, err
	}
	bs, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, err
	}
	if response.StatusCode != 200 {
		return -1, fmt.Errorf(string(bs))
	}
	var tt struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(bs, &tt); err != nil {
		return -1, err
	}
	if tt.Code != 0 {
		return -1, fmt.Errorf(tt.Msg)
	}
	return 0, nil
}
