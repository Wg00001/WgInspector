package task

import (
	"WgInspector/entities/config"
	"context"
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
