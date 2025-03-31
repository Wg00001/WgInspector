package websocket

import (
	"WgInspector/entities/client"
	client2 "WgInspector/usecase/client"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
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
	client2.RegisterDriver("websocket", &ClientWebSocket{})
}

const (
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

var _ client.Client = (*ClientWebSocket)(nil)

func (c *ClientWebSocket) Init(urlStr string) (_ client.Client, err error) {
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
		// 直接升级连接为 WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket 升级失败: %v", err)
			return
		}

		// 等待认证消息
		authenticated := make(chan bool, 1)
		go func() {
			defer close(authenticated)

			// 设置认证超时
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			// 读取认证消息
			var authMsg struct {
				Action    string `json:"action"`
				AuthToken string `json:"auth_token"`
			}

			err := conn.ReadJSON(&authMsg)
			if err != nil {
				log.Printf("读取认证消息失败: %v", err)
				conn.WriteJSON(map[string]interface{}{
					"action":  "authenticate_response",
					"success": false,
					"error":   "读取认证消息失败",
				})
				conn.Close()
				authenticated <- false
				return
			}

			// 验证消息类型
			if authMsg.Action != "authenticate" {
				log.Println("无效的认证消息类型")
				conn.WriteJSON(map[string]interface{}{
					"action":  "authenticate_response",
					"success": false,
					"error":   "无效的认证消息类型",
				})
				conn.Close()
				authenticated <- false
				return
			}

			// 解析 Basic 认证信息
			authType, authData, found := strings.Cut(authMsg.AuthToken, " ")
			if !found || authType != "Basic" {
				log.Println("认证格式错误")
				conn.WriteJSON(map[string]interface{}{
					"action":  "authenticate_response",
					"success": false,
					"error":   "认证格式错误",
				})
				conn.Close()
				authenticated <- false
				return
			}

			// 解码认证数据
			decoded, err := base64.StdEncoding.DecodeString(authData)
			if err != nil {
				log.Println("认证信息解码失败")
				conn.WriteJSON(map[string]interface{}{
					"action":  "authenticate_response",
					"success": false,
					"error":   "认证信息解码失败",
				})
				conn.Close()
				authenticated <- false
				return
			}

			username, password, found := strings.Cut(string(decoded), ":")
			if !found {
				log.Println("认证信息格式错误")
				conn.WriteJSON(map[string]interface{}{
					"action":  "authenticate_response",
					"success": false,
					"error":   "认证信息格式错误",
				})
				conn.Close()
				authenticated <- false
				return
			}

			// 验证用户身份
			user, err := client2.Auth(username, password)
			if err != nil {
				log.Printf("认证失败: %v", err)
				conn.WriteJSON(map[string]interface{}{
					"action":  "authenticate_response",
					"success": false,
					"error":   "认证失败",
				})
				conn.Close()
				authenticated <- false
				return
			}

			// 认证成功
			conn.WriteJSON(map[string]interface{}{
				"action":  "authenticate_response",
				"success": true,
			})

			// 重置读取超时
			conn.SetReadDeadline(time.Time{})

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

			authenticated <- true

			log.Printf("用户 %s WebSocket 连接认证成功", user.UserName)
		}()

		// 等待认证结果
		select {
		case success := <-authenticated:
			fmt.Println()
			if success {
				// 启动心跳检测
				go c.startPing(conn)
				// 处理 WebSocket 消息
				go c.handleWebSocketConnection(conn)
			}
		case <-time.After(10 * time.Second):
			log.Println("认证超时")
			conn.Close()
		}
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

func (c *ClientWebSocket) updateConnectionTime(conn *websocket.Conn) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if info, exists := c.conns[conn]; exists {
		info.lastActive = time.Now()
		info.timer.Reset(idleTimeout)
	}
}

func (c *ClientWebSocket) Close() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	for conn, info := range c.conns {
		info.timer.Stop()
		conn.Close()
	}
	c.conns = make(map[*websocket.Conn]*connInfo)
	return c.server.Close()
}

type Send struct {
	Action string `json:"action"`
	Type   string `json:"type"`
	Data   any    `json:"data"`
}

// UpdateCallback 服务端向客户端发送配置更新
func (c *ClientWebSocket) UpdateCallback(ctx context.Context, configType string, data any) error {
	marshal, err := json.Marshal(
		Send{
			Action: "config_update",
			Type:   configType,
			Data:   data,
		})
	if err != nil {
		return err
	}

	c.connMutex.RLock()
	defer c.connMutex.RUnlock()

	//排除当前链接
	var exclude *websocket.Conn
	if temp, ok := ctx.Value("exclude").(*websocket.Conn); ok {
		exclude = temp
	}

	for conn, info := range c.conns {
		if conn == exclude {
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, marshal)
		if err != nil {
			log.Printf("向用户 %s 发送消息失败: %v", info.username, err)
			continue
		}
	}
	return nil
}

func (c *ClientWebSocket) Listen(context.Context) {
	// 启动 HTTP 服务端（异步）
	go func() {
		if err := c.server.ListenAndServe(); err != nil {
			log.Printf("client: server Listen fail: %v\n", err)
		}
	}()
}
