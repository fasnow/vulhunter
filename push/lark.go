package push

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"cveHunter/config"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type LarkBot struct {
	webhookToken  string
	webhookSecret string
	HttpClient    *http.Client
}

var (
	larkBotPusher *LarkBot
	larkBotOnce   sync.Once
)

func GetLarkSingleton() *LarkBot {
	// 通过 sync.Once 确保仅执行一次实例化操作
	larkBotOnce.Do(func() {
		larkBotPusher = &LarkBot{
			webhookToken:  config.GetSingleton().LarkBot.WebHookAccessToken,
			webhookSecret: config.GetSingleton().LarkBot.WebHookSecret,
			HttpClient:    &http.Client{},
		}
	})
	return larkBotPusher
}

// Push 小于0表示推送失败
func (r *LarkBot) Push(msg string) (int, error) {
	if r.webhookToken == "" || r.webhookSecret == "" {
		return -1, fmt.Errorf("no webhook_access_token or webhook_secret")
	}

	// 获取当前时间戳（秒）,加签
	timestamp := strconv.FormatInt(time.Now().Unix()+3, 10)
	stringToSign := fmt.Sprintf("%s\n%s", timestamp, r.webhookSecret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, _ = h.Write([]byte{})
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	//https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot#%E6%94%AF%E6%8C%81%E5%8F%91%E9%80%81%E7%9A%84%E6%B6%88%E6%81%AF%E7%B1%BB%E5%9E%8B%E8%AF%B4%E6%98%8E
	//https://open.feishu.cn/document/common-capabilities/message-card/message-cards-content/card-structure/card-content
	t, _ := json.Marshal(map[string]any{
		"timestamp": timestamp,
		"sign":      sign,
		"msg_type":  "interactive",
		"card": map[string]any{
			"elements": []map[string]any{
				{
					"tag":     "markdown",
					"content": msg,
				},
			},
		},
	})
	request, err := http.NewRequest("POST", "https://open.feishu.cn/open-apis/bot/v2/hook/"+r.webhookToken, bytes.NewReader(t))
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
