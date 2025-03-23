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

func UpdateOrNew[T client.ConfigType](newConfig T) error {
	return nil
}

func GetConfigMeta() config2.ConfigMeta {
	config.RLock()
	defer config.RUnlock()
	return config.Meta
}
