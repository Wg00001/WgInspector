package websocket

import (
	"PgInspector/entities/client"
	"PgInspector/entities/config"
	client2 "PgInspector/usecase/client"
	config2 "PgInspector/usecase/config"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"reflect"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/22
 */

func init() {
	client2.RegisterDriver("websocket", ClientWebSocket{})
}

const (
	clientActionGet    = "config_get"
	clientActionSave   = "config_save"
	clientActionDelete = "config_delete"
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
	//go c.Listen()

	return c, nil
}

func (c ClientWebSocket) Close() error {
	return c.conn.Close()
}

// UpdateCallback 服务端向客户端发送配置更新
func (c ClientWebSocket) UpdateCallback(configType string, data any) error {
	fmt.Println(1)
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

func (c ClientWebSocket) Listen(ctx context.Context) {
	defer c.conn.Close()

	// 启动协程监听 context 取消事件
	go func() {
		<-ctx.Done()
		c.conn.Close() // 主动关闭连接，触发 ReadMessage 返回错误
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// 判断错误是否为 context 取消导致的正常关闭
			if ctx.Err() != nil {
				log.Printf("连接正常关闭: %v", ctx.Err())
			} else if websocket.IsUnexpectedCloseError(err) {
				log.Printf("连接异常关闭: %v", err)
			} else {
				log.Printf("读取消息错误: %v", err)
			}
			break // 退出循环，结束监听
		}

		// 解析和处理消息（原有逻辑）
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
			if err = c.handleConfigGet(); err != nil {
				log.Println("client - websocket: config_get handle err: " + err.Error())
			}
		case clientActionSave:
			if err = c.handleConfigSave(msg.ConfigType, msg.ConfigData); err != nil {
				log.Println("client - websocket: config_update handle err: " + err.Error())
			}
		default:
			log.Printf("client - websocket: 未知操作类型: %s\n", msg.Action)
		}
	}
}
func (c ClientWebSocket) handleConfigGet() error {
	mt := client2.GetConfigMeta()
	return c.conn.WriteJSON(mt)
}

func parseJson[T config.ConfigType](configData json.RawMessage) (T, error) {
	var res T
	err := json.Unmarshal(configData, &res)
	if err != nil {
		return res, fmt.Errorf("client - json parse fail: type %s, data: %v", reflect.TypeOf(res), configData)
	}
	return res, nil
}

// 客户端向服务端发送配置更新
func (c ClientWebSocket) handleConfigSave(configType string, configData json.RawMessage) (err error) {
	//cfg, err := jsonParse(configType, configData)
	switch configType {
	case config.TypeDB:
		return save(parseJson[config.DBConfig](configData))
	case config.TypeLog:
		return save(parseJson[config.LogConfig](configData))
	case config.TypeAlert:
		return save(parseJson[config.AlertConfig](configData))
	case config.TypeTask:
		return save(parseJson[config.TaskConfig](configData))
	case config.TypeAgent:
		return save(parseJson[config.AgentConfig](configData))
	case config.TypeAgentTask:
		return save(parseJson[config.AgentTaskConfig](configData))
	case config.TypeKBase:
		return save(parseJson[config.KnowledgeBaseConfig](configData))
	default:
		return fmt.Errorf("client - websocket: handle config save fail: type of configData not suppose: %s", configType)
	}
}

func save[T config.ConfigType](arg T, err error) error {
	if err != nil {
		return err
	}
	err = config2.Save(arg)
	if err != nil {
		return err
	}
	return client2.GetMetaOfType(arg)
}
