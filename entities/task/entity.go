package task

import (
	"WgInspector/entities/config"
	"context"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/25
 */

// Task 任务接口，巡检任务和ai分析任务实现此接口，用于交给cron分析。
type Task interface {
	Do(ctx context.Context) error
	GetCron() *config.Cron
	Identity() config.Identity
}

type Cron interface {
	Init() error
	AddTask(task Task) error
	Start()
	Exit()
	Monitor(ctx context.Context) (<-chan []Stat, error)
	DoNow(string) error
}

type Stat struct {
	UUID      string
	TaskName  string
	NextStart time.Time
	LastStart time.Time
	//Error     error
}
