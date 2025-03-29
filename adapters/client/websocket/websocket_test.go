package websocket

import (
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestClientWebsocketStart(t *testing.T) {
	err := client.Use(config.InitConfig{
		ClientDriver: "websocket",
		ClientURL:    "ws://127.0.0.1:9999",
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Listen(context.Background())
	defer client.Close()
	select {}
}

// 测试ClientWebSocket的Init方法
func TestClientWebSocket_Init(t *testing.T) {
	// 创建测试WebSocket服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
	}))
	defer ts.Close()

	// 转换为WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// 初始化客户端
	c, err := ClientWebSocket{}.Init(wsURL)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer c.Close()
}

// 测试处理config_get消息
func TestHandleConfigGet(t *testing.T) {
	ready := make(chan struct{})
	done := make(chan struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		// 等待客户端就绪
		<-ready

		// 发送请求
		req := map[string]string{"action": clientActionGet}
		if err := conn.WriteJSON(req); err != nil {
			t.Fatal(err)
		}

		// 读取响应
		var resp map[string]interface{}
		if err = conn.ReadJSON(&resp); err != nil {
			t.Fatal(err)
		}
		fmt.Println(resp)
		close(done)
	}))
	defer ts.Close()

	// 初始化客户端
	c, err := ClientWebSocket{}.Init("ws" + strings.TrimPrefix(ts.URL, "http"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 启动监听并通知就绪
	go func() {
		go c.Listen(context.Background())
		close(ready)
	}()

	// 等待测试完成或超时
	select {
	case <-done:
	}
}

// 测试处理config_update消息（DBConfig类型）
func TestHandleConfigUpdate_DB(t *testing.T) {
	ready := make(chan struct{}) // 同步客户端就绪
	done := make(chan struct{})  // 通知测试完成

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		// 等待客户端就绪
		<-ready

		// 发送config_update请求
		newDBConfig := config.DBConfig{Driver: "new updated"}
		req := map[string]interface{}{
			"action":      clientActionSave,
			"config_type": config.TypeDB,
			"config_data": newDBConfig,
		}
		if err := conn.WriteJSON(req); err != nil {
			t.Fatal(err)
		}

		// 读取更新回调
		var update interface{}
		if err := conn.ReadJSON(&update); err != nil {
			t.Fatal(err)
		}
		fmt.Println(update)
		close(done) // 通知测试完成
	}))
	defer ts.Close()

	// 初始化客户端
	err := client.Use(config.InitConfig{
		ClientDriver: "websocket",
		ClientURL:    "ws" + strings.TrimPrefix(ts.URL, "http"),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 启动监听并通知就绪
	client.Listen(context.Background())
	defer client.Close()
	close(ready) // 客户端开始监听后通知服务端

	// 等待测试完成或超时
	select {
	case <-done:
		//case <-time.After(5 * time.Second):
		//	t.Fatal("Test timed out")
	}
}

// 测试UpdateCallback方法
func TestUpdateCallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		// 读取客户端发送的更新
		var update struct {
			Type string          `json:"type"`
			Data config.DBConfig `json:"data"`
		}
		if err := conn.ReadJSON(&update); err != nil {
			t.Fatal(err)
		}

		if update.Type != config.TypeDB {
			t.Errorf("Expected type '%s', got '%s'", config.TypeDB, update.Type)
		}
		if update.Data.Driver != "callbackhost" {
			t.Errorf("Expected driver 'callbackhost', got '%s'", update.Data.Driver)
		}
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	c, err := ClientWebSocket{}.Init(wsURL)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer c.Close()

	// 触发UpdateCallback
	err = c.UpdateCallback(config.TypeDB, config.DBConfig{Driver: "callbackhost"})
	if err != nil {
		t.Fatalf("UpdateCallback failed: %v", err)
	}

	// 等待消息发送
	time.Sleep(100 * time.Millisecond)
}
