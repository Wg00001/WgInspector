package websocket

import (
	"PgInspector/entities/client"
	"PgInspector/entities/config"
	client2 "PgInspector/usecase/client"
	config2 "PgInspector/usecase/config"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/22
 */

const (
	clientActionGet    = "config_get"
	clientActionUpdate = "config_update"
)

type ClientWebSocket struct {
	conn     *websocket.Conn
	callback func(string, interface{})
}

var _ client.Client = (*ClientWebSocket)(nil)

func (c ClientWebSocket) Init(url string) (_ client.Client, err error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	c.conn = conn

	//启动消息监听协程
	go c.Listen()

	return c, nil
}

// UpdateCallback 服务端向客户端发送配置更新
func (c ClientWebSocket) UpdateCallback(configType string, data any) error {
	if err := c.conn.WriteJSON(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: configType,
		Data: data,
	}); err != nil {
		return fmt.Errorf("client - websocket: 发送失败: %w", err)
	}
	return nil
}

func (c ClientWebSocket) Listen() {
	defer c.conn.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("连接异常关闭: %v", err)
			}
			break
		}

		// 解析基础消息结构
		var msg struct {
			Action     string          `json:"action"`
			ConfigType string          `json:"config_type,omitempty"`
			ConfigData json.RawMessage `json:"config_data,omitempty"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		switch msg.Action {
		case clientActionGet:
			err := c.handleConfigGet()
			if err != nil {
				log.Println("client - websocket: config_get handle err: " + err.Error())
			}
		case clientActionUpdate:
			err := c.handleConfigUpdate(msg.ConfigType, msg.ConfigData)
			if err != nil {
				log.Println("client - websocket: config_update handle err: " + err.Error())
			}
		default:
			log.Printf("client - websocket: 未知操作类型: %s", msg.Action)
		}
	}
}

func (c ClientWebSocket) handleConfigGet() error {
	return c.conn.WriteJSON(client2.GetConfigMeta())
}

// todo: 测试
// 客户端向服务端发送配置更新
func (c ClientWebSocket) handleConfigUpdate(configType string, configData json.RawMessage) (err error) {
	switch configType {
	case client.ConfigTypeDB:
		var res config.DBConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.DBConfig))
	case client.ConfigTypeLog:
		var res config.LogConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.LogConfig))
	case client.ConfigTypeAlert:
		var res config.AlertConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.AlertConfig))

	case client.ConfigTypeTask:
		var res config.TaskConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.TaskConfig))

	case client.ConfigTypeAgent:
		var res config.AgentConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.AgentConfig))

	case client.ConfigTypeAgentTask:
		var res config.AgentTaskConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.AgentTaskConfig))

	case client.ConfigTypeKBase:
		var res config.KnowledgeBaseConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return update(any(res).(config.KnowledgeBaseConfig))
	default:
		return fmt.Errorf("client - websocket: handle config update fail: type of configData not suppose: %s", configType)
	}
}

func update[T client.ConfigType](data T) error {
	err := config2.Update(data)
	if err != nil {
		return err
	}
	//todo: test
	//return client2.SendUpdate(data)
	return client2.SendFullUpdate(data)
}
