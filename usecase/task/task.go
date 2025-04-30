package task

import (
	"WgInspector/entities/alerter"
	"WgInspector/entities/config"
	db2 "WgInspector/entities/db"
	"WgInspector/entities/logger"
	"WgInspector/entities/task"
	alerter2 "WgInspector/usecase/alerter"
	"WgInspector/usecase/db"
	logger2 "WgInspector/usecase/logger"
	"context"
	"fmt"
	"sync"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/1/19
 */

type Task struct {
	//Id string //批次编号, task每次启动会生成一个
	Config config.TaskConfig
}

type taskPlan struct {
	targetDBs []*config.DBConfig
	inspNodes []*config.InspConfig
}

var _ task.Task = (*Task)(nil)

func (t *Task) Do(ctx context.Context) error {
	taskId := time.Now().Format("20060102_150405")
	tp, err := newTaskPlan(t.Config)
	if err != nil {
		return err
	}
	fmt.Printf("task: start - %s\n", taskId)

	// 使用带缓冲的错误通道，缓冲区大小根据实际可能的最大错误数调整
	errChan := make(chan error, len(tp.inspNodes)*len(tp.targetDBs)*2)
	var wg sync.WaitGroup

	// 提取公共配置，避免在闭包中频繁访问 t.Config
	taskName := t.Config.Identity
	logID := t.Config.LogID
	alertID := t.Config.AlertID

	for _, inspect := range tp.inspNodes {
		inspName := inspect.Identity // 提取检查名称

		for _, tdb := range tp.targetDBs {
			if tdb == nil {
				continue
			}

			// 在每次迭代开始时检查上下文是否已取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			// 准备异步处理所需数据（避免在闭包中捕获循环变量）
			dbName := tdb.Identity.Name
			dbIdent := tdb.Identity // 假设这是需要传递的完整标识对象

			// 同步执行核心查询逻辑
			dbInstance := db.Get(tdb.Identity)
			query, err := dbInstance.Query(inspect.SQL)
			if err != nil {
				alerter2.GetAlert(alertID.Identity()).Send(alerter.Content{
					TimeStamp: time.Now(),
					Message:   fmt.Sprintf("insp_task get fail: query failed: %s", err),
					TaskName:  taskName,
					TaskID:    taskId,
					DBName:    dbIdent,
					InspName:  inspName,
				})
				continue
			}
			result, err := db2.RowsToResult(query)
			if err != nil {
				alerter2.GetAlert(alertID.Identity()).Send(alerter.Content{
					TimeStamp: time.Now(),
					Message:   fmt.Sprintf("insp_task rows fail:result conversion failed: %s", err),
					TaskName:  taskName,
					TaskID:    taskId,
					DBName:    dbIdent,
					InspName:  inspName,
				})
				continue
			}

			// 启动日志记录协程
			wg.Add(1)
			go func(content logger.LogContent) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						errChan <- fmt.Errorf("logging panic: %v", r)
					}
				}()

				if err := logger2.Get(logID.Identity()).Log(content); err != nil {
					errChan <- fmt.Errorf("logging failed: %w", err)
				}
			}(logger.LogContent{
				Timestamp: time.Now(),
				TaskName:  taskName.Name,
				TaskID:    taskId,
				InspName:  inspName.Name,
				DBName:    dbName,
				Result:    result.MarshallJSON(),
			})

			// 启动警报发送协程
			wg.Add(1)
			go func(content alerter.Content) { // 假设具体类型
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						errChan <- fmt.Errorf("alerting panic: %v", r)
					}
				}()
				if err := alerter2.GetAlert(alertID.Identity()).Send(content); err != nil {
					errChan <- fmt.Errorf("alert failed: %w", err)
				}
			}(alerter.Content{
				Success:   true,
				TimeStamp: time.Now(),
				TaskName:  taskName,
				TaskID:    taskId,
				DBName:    dbIdent,
				InspName:  inspName,
				Result:    result,
			})
		}
	}

	// 错误处理
	// 等待所有任务完成关闭通道
	wg.Wait()
	close(errChan)

	var errStr string
	// 处理剩余错误
	for err := range errChan {
		errStr += fmt.Sprintf("%v", err)
	}
	if errStr != "" {
		return fmt.Errorf("Async operation error: \n%s", errStr)
	}
	fmt.Printf("task submitted: %s\n", t.Config.Identity.Name)
	return nil
}

func (t *Task) GetCron() config.Cron {
	return t.Config.Cron
}

func (t *Task) Identity() config.Identity {
	return t.Config.Identity
}
