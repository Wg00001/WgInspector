package test

import (
	"WgInspector/adapters/start"
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"context"
	"testing"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/26
 */
func TestClientWebsocketStart(t *testing.T) {
	start.SetLocalConfigReaderOption("../../app/config", "local_file")
	start.InitOld()
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
