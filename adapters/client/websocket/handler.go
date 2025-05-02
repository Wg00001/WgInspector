package websocket

import (
	"WgInspector/entities/client"
	"WgInspector/entities/config"
	"WgInspector/usecase/agent"
	client2 "WgInspector/usecase/client"
	config2 "WgInspector/usecase/config"
	"WgInspector/usecase/task"
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
	clientActionGetID      = "config_get_id"
	clientActionUpdate     = "config_update"
	clientActionDelete     = "config_delete"
	clientActionCreate     = "config_create"
	clientActionChangePass = "change_password"
	clientNoticeConfirm    = "notice_confirm"
	clientNoticeGet        = "notice_get"
	clientTaskListen       = "task_listen"
	clientTaskClose        = "task_close"
	clientTaskDo           = "task_do"
	clientRefreshCron      = "task_cron_refresh"
)

type MsgMeta struct {
	Action     string `json:"action"`
	ConfigType string `json:"config_type,omitempty"`
}

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

func (c *ClientWebSocket) handleWebSocketConnection(conn *websocket.Conn, user client.User) {
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

		handleWithAuth := func(handleFunc HandleFunc, responseFunc ResponseFunc, authLevel int) error {
			if user.Level < authLevel {
				return response(conn, msg.MsgMeta, fmt.Errorf("no permission"))
			}
			return handler(conn, msg, handleFunc, responseFunc)
		}

		switch msg.Action {
		case clientActionGet:
			if msg.ConfigType == "Meta" || msg.ConfigType == "" {
				config2.RLock()
				logErr(clientActionGet, response(conn, msg.MsgMeta, config2.Meta))
				config2.RUnlock()
			} else {
				logErr(clientActionGet, handleWithAuth(getHandler, response, 0))
			}
		case clientActionGetID:
			logErr(clientActionGetID, handleWithAuth(getIdHandler, response, 0))
		case clientActionUpdate:
			logErr(clientActionUpdate, handleWithAuth(updateHandler, responseWithCallback, 1))
		case clientActionDelete:
			logErr(clientActionDelete, handleWithAuth(deleteHandler, responseWithCallback, 1))
		case clientActionCreate:
			logErr(clientActionCreate, handleWithAuth(createHandler, responseWithCallback, 1))
		case clientActionChangePass:
			logErr(clientActionChangePass, response(conn, MsgMeta{Action: clientActionChangePass}, c.handleChangePassword(conn, msg)))
		case clientNoticeConfirm:
			logErr(clientNoticeConfirm, handleNoticeConfirm(conn, msg, user))
		case clientNoticeGet:
			msg.ConfigData = message
			logErr(clientNoticeGet, handleNoticeGet(conn, msg))
		case clientTaskListen:
			taskCtxCancel()
			taskCtx, taskCtxCancel = context.WithCancel(context.Background())
			logErr(clientTaskListen, handleGetTaskStatus(taskCtx, conn, msg))
		case clientTaskClose:
			taskCtxCancel()
		case clientTaskDo:
			logErr(clientTaskDo, handleTaskDo(conn, msg))
		case clientRefreshCron:
			logErr(clientRefreshCron, handleRefreshCron(conn, msg))
		default:
			log.Printf("client websocket: 未知操作类型: %s\n", msg.Action)
		}
	}
}

type HandleFunc func(configType string, arg config.Id) any
type ResponseFunc func(conn *websocket.Conn, configType MsgMeta, obj any) error

func handler(
	conn *websocket.Conn,
	req RequestMsg,
	handleFunc HandleFunc,
	responseFunc ResponseFunc,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("websocket client - %v", r)
		}
	}()
	switch req.ConfigType {
	case config.TypeDB:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeDB, parseJson[config.DBConfig](req.ConfigData)))
	case config.TypeLog:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeLog, parseJson[config.LogConfig](req.ConfigData)))
	case config.TypeAlert:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeAlert, parseJson[config.AlertConfig](req.ConfigData)))
	case config.TypeTask:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeTask, parseJson[config.TaskConfig](req.ConfigData)))
	case config.TypeAgent:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeAgent, parseJson[config.AgentConfig](req.ConfigData)))
	case config.TypeAgentTask:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeAgentTask, parseJson[config.AgentTaskConfig](req.ConfigData)))
	case config.TypeKBase:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeKBase, parseJson[config.KnowledgeBaseConfig](req.ConfigData)))
	case config.TypeInspector:
		return responseFunc(conn, req.MsgMeta, handleFunc(config.TypeInspector, parseJson[config.InspConfig](req.ConfigData)))
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
	case string:
		return conn.WriteJSON(ResponseMsg{
			MsgMeta: msgMeta,
			Success: true,
			Message: t,
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

func parseJson[T config.Id](configData json.RawMessage) T {
	var res T
	if len(configData) == 0 {
		return res
	}
	err := json.Unmarshal(configData, &res)
	if err != nil {
		panic(fmt.Errorf("json parse fail: type %s, data: %v, err:%v", reflect.TypeOf(res), string(configData), err))
	}
	return res
}

func getHandler(configType string, arg config.Id) any {
	item, err := config2.GetMetaItem(configType)
	if err != nil {
		return err
	}
	return item
}

func getIdHandler(configType string, arg config.Id) any {
	res, err := config2.GetIdentityList(configType)
	if err != nil {
		return err
	}
	return res
}

func updateHandler(configType string, arg config.Id) any {
	err := config2.Save(config2.Key{
		ConfigType: configType,
		Identity:   arg.GetIdentity(),
	}, arg)
	if err != nil {
		return err
	}
	return getMetaItem(configType)
}

func deleteHandler(configType string, arg config.Id) any {
	err := config2.Del(config2.Key{
		ConfigType: configType,
		Identity:   arg.GetIdentity(),
	}, arg)
	if err != nil {
		return err
	}
	return getMetaItem(configType)
}

func createHandler(configType string, arg config.Id) any {
	err := config2.Save(config2.Key{
		ConfigType: configType,
		Identity:   arg.GetIdentity(),
	}, arg)
	if err != nil {
		return err
	}
	return getMetaItem(configType)
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

func handleNoticeConfirm(conn *websocket.Conn, msg RequestMsg, user client.User) error {
	var temp client.NoticeContent
	err := json.Unmarshal(msg.ConfigData, &temp)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	temp.UpdatedBy = user.UserName
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

func handleTaskDo(conn *websocket.Conn, msg RequestMsg) error {
	var uuid string
	err := json.Unmarshal(msg.ConfigData, &uuid)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	err = cron.DoNow(uuid)
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	return response(conn, msg.MsgMeta, "success")
}

func getMetaItem(dataType string) any {
	res, err := config2.GetMetaItem(dataType)
	if err != nil {
		return err
	}
	return res
}

func handleRefreshCron(conn *websocket.Conn, msg RequestMsg) error {
	err := cron.Init()
	if err != nil {
		response(conn, msg.MsgMeta, err)
		return err
	}
	for _, v := range config2.Meta.Tasks {
		t := task.NewInspTask(v)
		err := cron.AddTask(&t)
		if err != nil {
			response(conn, msg.MsgMeta, err)
			return err
		}
	}
	for _, v := range config2.Meta.AgentTasks {
		err := cron.AddTask(agent.NewTask(v))
		if err != nil {
			response(conn, msg.MsgMeta, err)
			return err
		}
	}
	return response(conn, msg.MsgMeta, "success")
}
