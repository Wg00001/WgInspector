package client

import (
	"PgInspector/entities/client"
	config2 "PgInspector/entities/config"
	"PgInspector/usecase/config"
	"fmt"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/20
 */

//客户端中心。
//根据设置在此注册客户端连接方式，例如websocket、http
//1. 客户端通过鉴权验证后，根据请求，系统一开始向客户端发送整个配置文件
//2. 此后每次配置修改时，调用回调函数将新版本发送给客户端

var (
	cli client.Client
	mu  sync.Mutex
)

func Register(client client.Client) error {
	mu.Lock()
	defer mu.Unlock()
	if cli != nil {
		return fmt.Errorf("client has been exist")
	}
	cli = client
	return nil
}

func SendUpdate[T client.ConfigType](data T) error {
	name, err := client.GetConfigTypeName(&data)
	if err != nil {
		return err
	}
	return cli.UpdateCallback(name, data)
}

func SendFullUpdate[T client.ConfigType](data T) error {
	config.RLock()
	defer config.RUnlock()
	switch any(data).(type) {
	case config2.DBConfig:
		return cli.UpdateCallback(client.ConfigTypeDB, GetConfigMeta().DBs)
	case config2.LogConfig:
		return cli.UpdateCallback(client.ConfigTypeLog, GetConfigMeta().Logs)
	case config2.AlertConfig:
		return cli.UpdateCallback(client.ConfigTypeAlert, GetConfigMeta().Alerts)
	case config2.TaskConfig:
		return cli.UpdateCallback(client.ConfigTypeTask, GetConfigMeta().Tasks)
	case config2.AgentConfig:
		return cli.UpdateCallback(client.ConfigTypeAgent, GetConfigMeta().Agent)
	case config2.AgentTaskConfig:
		return cli.UpdateCallback(client.ConfigTypeAgentTask, GetConfigMeta().AgentTasks)
	case config2.KnowledgeBaseConfig:
		return cli.UpdateCallback(client.ConfigTypeKBase, GetConfigMeta().KnowledgeBases)
	case config2.InspTree:
		return cli.UpdateCallback(client.ConfigTypeInspector, GetConfigMeta().Insp)
	default:
		return fmt.Errorf("unknown config type: %T", data)
	}
}

func UpdateOrNew[T client.ConfigType](newConfig T) error {
	return nil
}

func GetConfigMeta() config2.ConfigMeta {
	config.RLock()
	defer config.RUnlock()
	return config.Meta
}
