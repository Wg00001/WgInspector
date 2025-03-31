package client

import (
	"WgInspector/entities/client"
	config2 "WgInspector/entities/config"
	"WgInspector/usecase/config"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/20
 */

//客户端中心。
//根据设置在此注册客户端连接方式，例如websocket、http。设置好客户端后，全局只有使用该客户端才能连接上服务器。
//1. 客户端通过鉴权验证后，根据请求，系统一开始向客户端发送整个配置文件
//2. 此后每次配置修改时，调用回调函数将新版本发送给客户端

var (
	cli client.Client
	mu  sync.Mutex
)

func Listen(ctx context.Context) {
	cli.Listen(ctx)
	log.Println("client: server start Listen...")
}

func Close() error {
	return cli.Close()
}

func Register(client client.Client) error {
	mu.Lock()
	defer mu.Unlock()
	//if cli != nil {
	//	return fmt.Errorf("client has been exist")
	//}
	cli = client
	return nil
}

func CallBack(configType string, data any, omit *websocket.Conn) error {
	return cli.UpdateCallback(context.WithValue(context.Background(), "exclude", omit), configType, data)
}

func GetResponseMeta[T config2.Id](data T) any {
	config.RLock()
	defer config.RUnlock()
	switch any(data).(type) {
	case config2.DBConfig:
		return sliceCopy(config.Meta.DBs)
	case config2.LogConfig:
		return sliceCopy(config.Meta.Logs)
	case config2.AlertConfig:
		return sliceCopy(config.Meta.Alerts)
	case config2.TaskConfig:
		return sliceCopy(config.Meta.Tasks)
	case config2.AgentConfig:
		return config.Meta.Agent
	case config2.AgentTaskConfig:
		return sliceCopy(config.Meta.AgentTasks)
	case config2.KnowledgeBaseConfig:
		return sliceCopy(config.Meta.KnowledgeBases)
	case config2.InspTree:
		return config.Meta.Insp
	default:
		return fmt.Errorf("unknown config type: %T", data)
	}
}

func GetConfigMeta() config2.ConfigMeta {
	config.RLock()
	defer config.RUnlock()
	return config.Meta
}

func sliceCopy[T any](arr []T) []T {
	res := make([]T, len(arr))
	copy(res, arr)
	return res
}
