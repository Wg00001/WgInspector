package feishu

import (
	"WgInspector/entities/alerter"
	"WgInspector/entities/config"
	alerter2 "WgInspector/usecase/alerter"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/18
 */

func init() {
	alerter2.RegisterDriver("feishu", AlerterFeishu{})
}

type AlerterFeishu struct {
	config  config.AlertConfig
	WebHook string
}

func (a AlerterFeishu) Init(config config.AlertConfig) (alerter.Alerter, error) {
	a.config = config
	if webhook, ok := a.config.Header["webhook"]; !ok {
		return AlerterFeishu{}, fmt.Errorf("alerter init fail: config without field 'WebHook'")
	} else {
		a.WebHook = webhook
	}
	return a, nil
}

func (a AlerterFeishu) Send(content alerter.Content) error {
	// 将content编码为JSON
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content to JSON: %w", err)
	}

	// 创建HTTP POST请求
	req, err := http.NewRequest("POST", a.WebHook, bytes.NewBuffer(jsonContent))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, body)
	}

	return nil
}

var _ alerter.Alerter = (*AlerterFeishu)(nil)
