package client

import (
	"WgInspector/entities/client"
	config2 "WgInspector/entities/config"
	"WgInspector/usecase/config"
	"context"
	"fmt"
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

func SendUpdate[T config2.ConfigType](data T) error {
	name, err := config2.GetConfigTypeName(&data)
	if err != nil {
		return err
	}
	return cli.UpdateCallback(name, data)
}

func GetMetaOfType[T config2.ConfigType](data T) error {
	meta := GetConfigMeta()
	switch any(data).(type) {
	case config2.DBConfig:
		return cli.UpdateCallback(config2.TypeDB, meta.DBs)
	case config2.LogConfig:
		return cli.UpdateCallback(config2.TypeLog, GetConfigMeta().Logs)
	case config2.AlertConfig:
		return cli.UpdateCallback(config2.TypeAlert, GetConfigMeta().Alerts)
	case config2.TaskConfig:
		return cli.UpdateCallback(config2.TypeTask, GetConfigMeta().Tasks)
	case config2.AgentConfig:
		return cli.UpdateCallback(config2.TypeAgent, GetConfigMeta().Agent)
	case config2.AgentTaskConfig:
		return cli.UpdateCallback(config2.TypeAgentTask, GetConfigMeta().AgentTasks)
	case config2.KnowledgeBaseConfig:
		return cli.UpdateCallback(config2.TypeKBase, GetConfigMeta().KnowledgeBases)
	case config2.InspTree:
		return cli.UpdateCallback(config2.TypeInspector, GetConfigMeta().Insp)
	default:
		return fmt.Errorf("unknown config type: %T", data)
	}
}

func UpdateOrNew[T config2.ConfigType](newConfig T) error {
	return nil
}

func GetConfigMeta() config2.ConfigMeta {
	config.RLock()
	defer config.RUnlock()
	return config.Meta
}
