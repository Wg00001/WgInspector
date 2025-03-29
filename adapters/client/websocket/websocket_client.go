package websocket

import (
	"WgInspector/entities/client"
	"WgInspector/entities/config"
	client2 "WgInspector/usecase/client"
	config2 "WgInspector/usecase/config"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"
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

	// 连接超时时间
	idleTimeout = 3 * time.Hour
	// ping 检查间隔
	pingPeriod = 30 * time.Second
)

type ClientWebSocket struct {
	server    *http.Server
	conns     map[*websocket.Conn]*connInfo
	connMutex sync.RWMutex
	parsedURL *url.URL
	callback  func(string, interface{})
}

type connInfo struct {
	conn       *websocket.Conn
	lastActive time.Time
	timer      *time.Timer
	username   string
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
	c.conns = make(map[*websocket.Conn]*connInfo)
	c.connMutex = sync.RWMutex{}

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
		// 获取认证信息
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "未提供认证信息", http.StatusUnauthorized)
			return
		}

		// 解析 Basic 认证信息
		authType, authData, found := strings.Cut(authHeader, " ")
		if !found || authType != "Basic" {
			http.Error(w, "认证格式错误", http.StatusBadRequest)
			return
		}

		// 解码认证数据
		decoded, err := base64.StdEncoding.DecodeString(authData)
		if err != nil {
			http.Error(w, "认证信息解码失败", http.StatusBadRequest)
			return
		}

		username, password, found := strings.Cut(string(decoded), ":")
		if !found {
			http.Error(w, "认证信息格式错误", http.StatusBadRequest)
			return
		}

		// 验证用户身份
		user, err := client2.Auth(username, password)
		if err != nil {
			log.Printf("认证失败: %v", err)
			http.Error(w, "认证失败", http.StatusUnauthorized)
			return
		}

		// 认证成功，升级连接为 WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}

		// 创建连接信息
		info := &connInfo{
			conn:       conn,
			lastActive: time.Now(),
			username:   user.UserName,
		}

		// 设置连接超时定时器
		info.timer = time.AfterFunc(idleTimeout, func() {
			c.closeConnection(conn, "连接超时")
		})

		// 保存连接
		c.connMutex.Lock()
		c.conns[conn] = info
		c.connMutex.Unlock()

		log.Printf("用户 %s WebSocket 连接成功", user.UserName)

		// 启动心跳检测
		go c.startPing(conn)

		// 处理 WebSocket 消息
		go c.handleWebSocketConnection(conn)
	})

	return c, nil
}

func (c *ClientWebSocket) startPing(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("发送 ping 失败: %v", err)
				c.closeConnection(conn, "ping 失败")
				return
			}
		}
	}
}

func (c *ClientWebSocket) closeConnection(conn *websocket.Conn, reason string) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if info, exists := c.conns[conn]; exists {
		log.Printf("关闭用户 %s 的连接: %s", info.username, reason)
		info.timer.Stop()
		conn.Close()
		delete(c.conns, conn)
	}
}

func (c *ClientWebSocket) handleWebSocketConnection(conn *websocket.Conn) {
	defer c.closeConnection(conn, "连接结束")

	// 设置消息处理函数
	conn.SetPongHandler(func(string) error {
		c.updateConnectionTime(conn)
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("连接异常关闭: %v", err)
			}
			return
		}

		// 更新最后活动时间
		c.updateConnectionTime(conn)

		// 解析和处理消息
		msg := MessageStruct{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		switch msg.Action {
		case clientActionGet:
			if err = c.handleConfigGet(conn); err != nil {
				log.Printf("处理 config_get 失败: %v", err)
			}
		case clientActionSave:
			if err = c.handleConfigSave(msg.ConfigType, msg.ConfigData); err != nil {
				log.Printf("处理 config_save 失败: %v", err)
			}
		case clientActionDelete:
			log.Printf("收到删除请求")
		default:
			log.Printf("未知操作类型: %s", msg.Action)
		}
	}
}

func (c *ClientWebSocket) updateConnectionTime(conn *websocket.Conn) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if info, exists := c.conns[conn]; exists {
		info.lastActive = time.Now()
		info.timer.Reset(idleTimeout)
	}
}

func (c ClientWebSocket) Close() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	for conn, info := range c.conns {
		info.timer.Stop()
		conn.Close()
	}
	c.conns = make(map[*websocket.Conn]*connInfo)
	return c.server.Close()
}

// UpdateCallback 服务端向客户端发送配置更新
func (c *ClientWebSocket) UpdateCallback(configType string, data any) error {
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

	c.connMutex.RLock()
	defer c.connMutex.RUnlock()

	for conn, info := range c.conns {
		err = conn.WriteMessage(websocket.TextMessage, marshal)
		if err != nil {
			log.Printf("向用户 %s 发送消息失败: %v", info.username, err)
			continue
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
