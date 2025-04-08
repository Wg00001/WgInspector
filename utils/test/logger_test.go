package test

import (
	"WgInspector/adapters/start"
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
	start.SetLocalConfigReaderOption("../../app/config", "yaml")
	start.Init()
	start.Run(context.TODO())
}
