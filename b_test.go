package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// 运行命令（带性能分析）：
// go test -bench=. -benchmem -benchtime=10s -cpuprofile=cpu.pprof -memprofile=mem.pprof

// 认证配置
const (
	authUser     = "admin"
	authPassword = "123456"
	wsPath       = "ws://0.0.0.0:9999/"
)

// 基准测试数据结构
type WSRequest struct {
	Action     string `json:"action"`
	ConfigType string `json:"config_type"`
}

func BenchmarkWebSocket(b *testing.B) {
	// 准备带超时的 Dialer
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second, // 握手超时
		Proxy:            http.ProxyFromEnvironment,
	}

	// 准备认证头
	authHeader := http.Header{}
	authHeader.Set("authenticate", "Basic "+
		base64.StdEncoding.EncodeToString([]byte(authUser+":"+authPassword)))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now()
			var conn *websocket.Conn
			var err error

			// 建立连接（带重试机制）
			for retry := 0; retry < 3; retry++ {
				conn, _, err = dialer.Dial(wsPath, authHeader)
				if err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			if err != nil {
				b.Errorf("连接失败: %v", err)
				continue
			}
			defer conn.Close()

			// 配置超时设置
			conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))

			// 发送请求
			req := WSRequest{
				Action:     "config_get",
				ConfigType: "Meta",
			}
			if err := conn.WriteJSON(req); err != nil {
				b.Errorf("发送失败: %v", err)
				continue
			}

			// 读取响应
			var response bytes.Buffer
			_, msg, err := conn.ReadMessage()
			if err != nil {
				b.Errorf("接收失败: %v", err)
				continue
			}
			response.Write(msg)

			// 验证响应
			if !bytes.Contains(response.Bytes(), []byte("success")) {
				b.Error("响应验证失败")
			}

			// 记录性能指标
			elapsed := time.Since(start)
			b.ReportMetric(float64(elapsed.Milliseconds()), "latency_ms")
			b.ReportMetric(1, "requests") // 统计总请求数
		}
	})

	// 输出附加指标
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
}
