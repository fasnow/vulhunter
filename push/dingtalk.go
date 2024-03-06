package push

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"cveHunter/config"
	"cveHunter/db"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DingTalk struct {
	webhookToken  string
	webhookSecret string
	HttpClient    *http.Client
}

var (
	dingtalkPusher *DingTalk
	dingtalkOnce   sync.Once
)

func GetDingTalkSingleton() *DingTalk {
	// 通过 sync.Once 确保仅执行一次实例化操作
	dingtalkOnce.Do(func() {
		dingtalkPusher = &DingTalk{
			webhookToken:  config.GetSingleton().DingTalk.WebHookAccessToken,
			webhookSecret: config.GetSingleton().DingTalk.WebHookSecret,
			HttpClient:    &http.Client{},
		}
	})
	return dingtalkPusher
}

// Push 小于0表示推送失败
func (r *DingTalk) Push(cves ...db.GithubCVE) (int, error) {
	if r.webhookToken == "" || r.webhookSecret == "" {
		return -1, fmt.Errorf("no webhook_access_token or webhook_secret")
	}
	// 获取当前时间戳（毫秒）,加签
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := fmt.Sprintf("%s\n%s", timestamp, r.webhookSecret)
	h := hmac.New(sha256.New, []byte(r.webhookSecret))
	_, _ = io.WriteString(h, stringToSign)
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	params := url.Values{
		"access_token": []string{r.webhookToken},
		"timestamp":    []string{timestamp},
		"sign":         []string{sign},
	}

	var items []string
	for _, cve := range cves {
		items = append(items,
			fmt.Sprintf("**漏洞编号**:%s  \n**地址**:%s  \n**描述**:  %s  \n", cve.Name, cve.HtmlUrl, cve.Description))
	}

	t, _ := json.Marshal(map[string]any{
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
	request, err := http.NewRequest("POST", "https://oapi.dingtalk.com/robot/send?"+params.Encode(), bytes.NewReader(t))
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
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(bs, &tt); err != nil {
		return -1, err
	}
	if tt.ErrCode != 0 {
		return -1, fmt.Errorf(tt.ErrMsg)
	}
	return 0, nil
}
