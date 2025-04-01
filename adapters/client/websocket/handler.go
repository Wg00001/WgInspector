package websocket

import (
	"WgInspector/entities/client"
	"WgInspector/entities/config"
	client2 "WgInspector/usecase/client"
	config2 "WgInspector/usecase/config"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"reflect"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/31
 */

const (
	clientActionGet        = "config_get"
	clientActionSave       = "config_save"
	clientActionDelete     = "config_delete"
	clientActionCreate     = "config_create"
	clientActionChangePass = "change_password"
)

type RequestMsg struct {
	MsgMeta
	ConfigData json.RawMessage `json:"config_data,omitempty"`
	OldPass    string          `json:"old_password,omitempty"`
	NewPass    string          `json:"new_password,omitempty"`
}

type ResponseMsg struct {
	MsgMeta
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ConfigData any    `json:"config_data"`
}

type MsgMeta struct {
	Action     string `json:"action"`
	ConfigType string `json:"config_type"`
}

func (c *ClientWebSocket) handleWebSocketConnection(conn *websocket.Conn) {
	defer c.closeConnection(conn, "连接结束")

	// 设置消息处理函数
	conn.SetPongHandler(func(string) error {
		c.updateConnectionTime(conn)
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("连接异常关闭: %v", err)
			}
			return
		}

		// 更新最后活动时间
		c.updateConnectionTime(conn)

		// 解析和处理消息
		msg := RequestMsg{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		switch msg.Action {
		case clientActionGet:
			if msg.ConfigType == "Meta" || msg.ConfigType == "" {
				config2.RLock()
				logErr(clientActionGet, response(conn, msg.MsgMeta, client2.GetConfigMeta()))
				config2.RUnlock()
			} else {
				logErr(clientActionGet, handler(conn, msg, getHandler, response))
			}
		case clientActionSave:
			logErr(clientActionSave, handler(conn, msg, saveHandler, responseWithCallback))
		case clientActionDelete:
			logErr(clientActionDelete, handler(conn, msg, deleteHandler, responseWithCallback))
		case clientActionCreate:
			logErr(clientActionCreate, handler(conn, msg, createHandler, responseWithCallback))
		case clientActionChangePass:
			logErr(clientActionChangePass, response(conn, MsgMeta{Action: clientActionChangePass}, c.handleChangePassword(conn, msg)))
		default:
			log.Printf("client websocket: 未知操作类型: %s\n", msg.Action)
		}
	}
}

type HandleFunc func(arg config.Id, err error) any
type ResponseFunc func(conn *websocket.Conn, configType MsgMeta, obj any) error

func handler(
	conn *websocket.Conn,
	req RequestMsg,
	handleFunc HandleFunc,
	responseFunc ResponseFunc,
) (err error) {
	switch req.ConfigType {
	case config.TypeDB:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.DBConfig](req.ConfigData)))
	case config.TypeLog:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.LogConfig](req.ConfigData)))
	case config.TypeAlert:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.AlertConfig](req.ConfigData)))
	case config.TypeTask:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.TaskConfig](req.ConfigData)))
	case config.TypeAgent:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.AgentConfig](req.ConfigData)))
	case config.TypeAgentTask:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.AgentTaskConfig](req.ConfigData)))
	case config.TypeKBase:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.KnowledgeBaseConfig](req.ConfigData)))
	case config.TypeInspector:
		return responseFunc(conn, req.MsgMeta, handleFunc(parseJson[config.InspNode](req.ConfigData)))
	default:
		return fmt.Errorf("client - websocket: handle fail: type of configData not suppose: %s", req.ConfigType)
	}
}

func response(conn *websocket.Conn, msgMeta MsgMeta, obj any) error {
	switch t := obj.(type) {
	case error:
		return conn.WriteJSON(ResponseMsg{
			MsgMeta:    msgMeta,
			Success:    false,
			Message:    t.Error(),
			ConfigData: t,
		})
	default:
		return conn.WriteJSON(ResponseMsg{
			MsgMeta:    msgMeta,
			Success:    true,
			ConfigData: obj,
		})
	}
}

func responseWithCallback(conn *websocket.Conn, msgMeta MsgMeta, obj any) error {
	err := response(conn, msgMeta, obj)
	if err != nil {
		return err
	}
	return client2.CallBack(msgMeta.ConfigType, obj, conn)
}

func parseJson[T config.Id](configData json.RawMessage) (T, error) {
	var res T
	err := json.Unmarshal(configData, &res)
	if err != nil {
		return res, fmt.Errorf("client - json parse fail: type %s, data: %v", reflect.TypeOf(res), string(configData))
	}
	return res, nil
}

func getHandler(arg config.Id, err error) any {
	if err != nil {
		return err
	}
	return client2.GetResponseMeta(arg)
}

func saveHandler(arg config.Id, err error) any {
	if err != nil {
		return err
	}
	err = config2.Save(arg)
	if err != nil {
		return err
	}
	return client2.GetResponseMeta(arg)
}

func deleteHandler(arg config.Id, err error) any {
	if err != nil {
		return err
	}
	err = config2.Del(arg)
	if err != nil {
		return err
	}
	return client2.GetResponseMeta(arg)
}

func createHandler(arg config.Id, err error) any {
	if err != nil {
		return err
	}
	err = config2.AppendConfigs(arg)
	if err != nil {
		return err
	}
	return client2.GetResponseMeta(arg)
}

func (c *ClientWebSocket) handleChangePassword(conn *websocket.Conn, msg RequestMsg) error {
	// 获取当前连接的用户信息
	c.connMutex.RLock()
	info, exists := c.conns[conn]
	c.connMutex.RUnlock()
	if !exists {
		return fmt.Errorf("连接信息不存在")
	}

	// 验证旧密码
	_, err := client2.Auth(info.username, msg.OldPass)
	if err != nil {
		return fmt.Errorf("旧密码验证失败: %w", err)
	}

	// 更新密码
	err = client2.UpdateUser(client.User{
		UserName: info.username,
		Password: msg.NewPass,
	})
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

func logErr(action string, err error) {
	if err != nil {
		log.Printf("handle action '%s' fail: %s", action, err)
	}
}
