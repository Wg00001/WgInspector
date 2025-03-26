package test

import (
	"PgInspector/adapters/start"
	"PgInspector/entities/config"
	"PgInspector/usecase/client"
	"context"
	"testing"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/26
 */
func TestClientWebsocketStart(t *testing.T) {
	start.SetConfigPath("../../app/config", "local_file")
	start.Init()
	err := client.Use(config.DefaultConfig{
		ClientDriver: "websocket",
		ClientURL:    "ws://127.0.0.1:9999",
	})
	if err != nil {
		t.Fatal(err)
	}
	closeFunc := client.Listen(context.Background())
	defer closeFunc()
	select {}
}
