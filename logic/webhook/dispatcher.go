package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/public/common"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Fire 是事件触发入口——异步、非阻塞。
// 无论连接器有没有配置 webhook，都可以安全调用此函数。
// 它会在后台 goroutine 中加载 webhook 列表并逐个发送。
func Fire(event EventType, connectorID uint, provider string, user *model.User) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				common.Log.Errorf("[webhook] panic recovered in Fire: %v", r)
			}
		}()

		hooks, err := loadWebhooks(connectorID, string(event))
		if err != nil {
			common.Log.Warnf("[webhook] load webhooks failed for connector %d: %v", connectorID, err)
			return
		}
		if len(hooks) == 0 {
			return
		}

		// 逐个弹出并异步发送，不阻塞后面的 webhook
		// while(!empty(webhooks)){ async hook(webhooks.pop()) }
		for len(hooks) > 0 {
			hook := hooks[0]
			hooks = hooks[1:] // pop

			go func(h model.OAuthWebhook) {
				defer func() {
					if r := recover(); r != nil {
						common.Log.Errorf("[webhook] panic in sender: %v", r)
					}
				}()

				var fields []string
				_ = json.Unmarshal(h.Fields, &fields)

				payload := Payload{
					Event:       string(event),
					ConnectorID: connectorID,
					Provider:    provider,
					Timestamp:   time.Now().Unix(),
					User:        BuildUserFields(user, fields),
				}

				if err := send(h.URL, payload); err != nil {
					common.Log.Warnf("[webhook] send to %s failed: %v", h.URL, err)
				} else {
					common.Log.Infof("[webhook] sent %s to %s OK", event, h.URL)
				}
			}(hook)
		}
	}()
}

// loadWebhooks 从数据库加载指定连接器下、订阅了指定事件的、已启用的 webhook 列表
func loadWebhooks(connectorID uint, event string) ([]model.OAuthWebhook, error) {
	var all []model.OAuthWebhook
	err := common.DB.
		Where("oauth_connector_id = ? AND enabled = ?", connectorID, true).
		Find(&all).Error
	if err != nil {
		return nil, err
	}

	// 过滤出订阅了该事件的 webhook
	var matched []model.OAuthWebhook
	for _, h := range all {
		var events []string
		if err := json.Unmarshal(h.Events, &events); err != nil {
			continue
		}
		for _, e := range events {
			if e == event {
				matched = append(matched, h)
				break
			}
		}
	}
	return matched, nil
}

// send 发送 POST 请求到目标 URL
func send(url string, payload Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-ldap-admin-webhook")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body) // drain body

	if resp.StatusCode >= 300 {
		return fmt.Errorf("http status %d", resp.StatusCode)
	}
	return nil
}
