package test

import (
	start2 "WgInspector/app/start"
	"context"
	"testing"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/15
 */

func TestLogger(t *testing.T) {
	//cron.Init()
	//cron.AddTask(task.Get("task1"))
	//cron.Start()
	select {}
}

func TestStart(t *testing.T) {
	start2.SetLocalConfigReaderOption("../../app/config", "yaml")
	start2.InitOld()
	start2.Run(context.TODO())
}
