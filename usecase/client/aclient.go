package client

import (
	"WgInspector/entities/client"
	"context"
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

func Notice(content client.NoticeContent) error {
	//1. 持久化
	err := CreateNotice(content)
	if err != nil {
		return err
	}
	//2. 发送给用户客户端
	mu.Lock()
	defer mu.Unlock()
	return cli.Notice(content)
}
