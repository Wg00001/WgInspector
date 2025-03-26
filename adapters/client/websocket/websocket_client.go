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
	"net/http"
	"net/url"
	"reflect"
	"strings"
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
	server    *http.Server
	conns     map[*websocket.Conn]bool
	parsedURL *url.URL

	callback func(string, interface{})
}

type MessageStruct struct {
	Action     string          `json:"action"`
	ConfigType string          `json:"config_type,omitempty"`
	ConfigData json.RawMessage `json:"config_data,omitempty"`
}

var _ client.Client = (*ClientWebSocket)(nil)

func (c ClientWebSocket) Init(urlStr string) (_ client.Client, err error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	c.conns = make(map[*websocket.Conn]bool)
	// 解析 URL 获取监听地址和端口
	c.parsedURL, err = url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("URL 解析失败: %w", err)
	}
	if !strings.HasPrefix(c.parsedURL.Path, "/") {
		c.parsedURL.Path = "/" + c.parsedURL.Path
	}
	c.server = &http.Server{
		Addr: c.parsedURL.Host,
	}

	// 设置 HTTP 路由和处理函数
	http.HandleFunc(c.parsedURL.Path, func(w http.ResponseWriter, r *http.Request) {
		// 升级 HTTP 连接为 WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("WebSocket 升级失败: %v\n", err)
			return
		}

		// 保存连接
		c.conns[conn] = true
		log.Println("websocket connected")
		//进行处理
		go func() {
			defer conn.Close()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					// 判断错误是否为 context 取消导致的正常关闭
					if websocket.IsUnexpectedCloseError(err) {
						log.Printf("连接异常关闭: %v", err)
					} else {
						log.Printf("读取消息错误: %v", err)
					}
					break // 退出循环，结束监听
				}

				// 解析和处理消息
				msg := MessageStruct{}
				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("消息解析失败: %v", err)
					continue
				}

				switch msg.Action {
				case clientActionGet:
					if err = c.handleConfigGet(conn); err != nil {
						log.Println("client - websocket: config_get handle err: " + err.Error())
					}
				case clientActionSave:
					if err = c.handleConfigSave(msg.ConfigType, msg.ConfigData); err != nil {
						log.Println("client - websocket: config_update handle err: " + err.Error())
					}
				case clientActionDelete:

				default:
					log.Printf("client - websocket: 未知操作类型: %s\n", msg.Action)
				}
			}
		}()
	})

	return c, nil
}

func (c ClientWebSocket) Close() error {
	return c.server.Close()
}

// UpdateCallback 服务端向客户端发送配置更新
func (c ClientWebSocket) UpdateCallback(configType string, data any) error {
	marshal, err := json.Marshal(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: configType,
		Data: data,
	})
	if err != nil {
		return err
	}
	for conn, ok := range c.conns {
		if ok {
			err = conn.WriteMessage(websocket.TextMessage, marshal)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c ClientWebSocket) Listen(context.Context) {
	// 启动 HTTP 服务端（异步）
	go func() {
		if err := c.server.ListenAndServe(); err != nil {
			log.Printf("client: server Listen fail: %v\n", err)
		}
	}()

}

func (c ClientWebSocket) handleConfigGet(conn *websocket.Conn) error {
	mt := client2.GetConfigMeta()
	return conn.WriteJSON(mt)
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
