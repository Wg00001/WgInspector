package websocket

import (
	"WgInspector/entities/client"
	"WgInspector/entities/config"
	"WgInspector/usecase/agent"
	client2 "WgInspector/usecase/client"
	config2 "WgInspector/usecase/config"
	"WgInspector/usecase/task/cron"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"reflect"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/31
 */

const (
	clientActionGet        = "config_get"
	clientActionUpdate     = "config_update"
	clientActionDelete     = "config_delete"
	clientActionCreate     = "config_create"
	clientActionChangePass = "change_password"
	clientNoticeConfirm    = "notice_confirm"
	clientNoticeGet        = "notice_get"
	clientTaskListen       = "task_listen"
	clientTaskClose        = "task_close"
)

// 请求过来的数据的格式
type RequestMsg struct {
	MsgMeta
	ConfigData json.RawMessage `json:"config_data,omitempty"`
	OldPass    string          `json:"old_password,omitempty"`
	NewPass    string          `json:"new_password,omitempty"`
	Confirm    bool            `json:"confirm,omitempty"`
}

// 发送过去的响应的格式
type ResponseMsg struct {
	MsgMeta
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ConfigData any    `json:"config_data"`
}

type MsgMeta struct {
	Action     string `json:"action"`
	ConfigType string `json:"config_type,omitempty"`
}

func (c *ClientWebSocket) handleWebSocketConnection(conn *websocket.Conn) {
	defer c.closeConnection(conn, "连接结束")

	// 设置消息处理函数
	conn.SetPongHandler(func(string) error {
		c.updateConnectionTime(conn)
		return nil
	})
	var taskCtx context.Context
	taskCtxCancel := func() {}
	defer func() {
		taskCtxCancel()
	}()
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
		case clientActionUpdate:
			logErr(clientActionUpdate, handler(conn, msg, updateHandler, responseWithCallback), config2.SaveConfig(msg.ConfigType))
		case clientActionDelete:
			logErr(clientActionDelete, handler(conn, msg, deleteHandler, responseWithCallback), config2.SaveConfig(msg.ConfigType))
		case clientActionCreate:
			logErr(clientActionCreate, handler(conn, msg, createHandler, responseWithCallback), config2.SaveConfig(msg.ConfigType))
		case clientActionChangePass:
			logErr(clientActionChangePass, response(conn, MsgMeta{Action: clientActionChangePass}, c.handleChangePassword(conn, msg)))
		case clientNoticeConfirm:
			logErr(clientNoticeConfirm, handleNoticeConfirm(conn, msg))
		case clientNoticeGet:
			msg.ConfigData = message
			logErr(clientNoticeGet, handleNoticeGet(conn, msg))
		case clientTaskListen:
			taskCtxCancel()
			taskCtx, taskCtxCancel = context.WithCancel(context.Background())
			logErr(clientTaskListen, handleGetTaskStatus(taskCtx, conn, msg))
		case clientTaskClose:
			taskCtxCancel()
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

var writeMu sync.Mutex

func response(conn *websocket.Conn, msgMeta MsgMeta, obj any) error {
	writeMu.Lock()
	defer writeMu.Unlock()
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

func updateHandler(arg config.Id, err error) any {
	if err != nil {
		return err
	}
	err = config2.Set(arg)
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

func logErr(action string, errs ...error) {
	for _, e := range errs {
		if e != nil {
			log.Printf("handle action '%s' fail: %s", action, errs)
			return
		}
	}
}

func handleNoticeConfirm(conn *websocket.Conn, msg RequestMsg) error {
	var temp client.NoticeContent
	err := json.Unmarshal(msg.ConfigData, &temp)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	if msg.Confirm {
		err = agent.KBaseSave(temp.Content)
		if err != nil {
			response(conn, msg.MsgMeta, err)
			return err
		}
	}
	err = client2.UpdateNotice(temp)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	//若要提高性能，可以客户端只返回id，服务端用id读取更新数据库并保存入知识库中。
	return response(conn, msg.MsgMeta, "success")
}

func handleNoticeGet(conn *websocket.Conn, msg RequestMsg) error {
	pages := struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	}{}
	err := json.Unmarshal(msg.ConfigData, &pages)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	notices, err := client2.GetNotice(pages.Page, pages.PageSize)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	return response(conn, msg.MsgMeta, notices)
}

func handleGetTaskStatus(parentCtx context.Context, conn *websocket.Conn, msg RequestMsg) error {
	ctx, f := context.WithCancel(parentCtx)
	ch, err := cron.Monitor(ctx)
	if err != nil {
		f()
		return err
	}
	go func() {
		defer f() //取消context
		for {
			select {
			case <-parentCtx.Done():
				return
			default:
			}
			status, ok := <-ch
			if !ok {
				return
			}
			err := response(conn, msg.MsgMeta, status)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
	return nil
}
