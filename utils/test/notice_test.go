package test

import (
	start2 "WgInspector/app/start"
	client2 "WgInspector/entities/client"
	"WgInspector/usecase/client"
	"context"
	"testing"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/6
 */

func TestNotice(t *testing.T) {
	start2.SetLocalConfigReaderOption("../../app/config", "yaml")
	start2.InitOld()
	//start.Run(context.TODO())
	t.Run("send", func(t *testing.T) {
		client.CreateNotice(client2.NoticeContent{
			Content:     "content notice test1",
			Time:        time.Now(),
			ConfirmStat: client2.Unread,
		})
		client.CreateNotice(client2.NoticeContent{
			Content:     "content notice test2",
			Time:        time.Now(),
			ConfirmStat: client2.Allow,
		})
		client.CreateNotice(client2.NoticeContent{
			Content:     "content notice test3",
			Time:        time.Now(),
			ConfirmStat: client2.UnConfirm,
		})
	})

	t.Run("get", func(t *testing.T) {
		client.Listen(context.TODO())

	})
}
