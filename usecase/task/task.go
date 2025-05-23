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
	"github.com/wg00001/wgo-sdk/wg"
	"strings"
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

	// 创建工作池
	workerCount := 5 // 可以根据需要调整
	jobChan := make(chan struct {
		inspect *config.InspConfig
		db      *config.DBConfig
	}, workerCount)
	errChan := make(chan error, len(tp.inspNodes)*len(tp.targetDBs)*2+24)
	contentChan := make(chan logger.LogContent, len(tp.inspNodes)*len(tp.targetDBs)+8)
	var wgroup sync.WaitGroup

	// 提取公共配置
	taskName := t.Config.Identity
	alertID := t.Config.AlertID

	dbList, err := db.ConnList(t.Config.TargetDB)
	if err != nil {
		alerter2.GetAlert(alertID.Identity()).Send(alerter.Content{
			TimeStamp: time.Now(),
			Message:   fmt.Sprintf("任务运行出现错误: 数据库连接失败\n%s\n", err),
			TaskName:  taskName,
			TaskID:    taskId,
		})
		return err
	}
	dbIdx := wg.SliceToIndex(dbList, func(item *db2.SqlDB) config.Identity {
		return item.Config.Identity
	})

	// 启动工作池
	for i := 0; i < workerCount; i++ {
		wgroup.Add(1)
		go func() {
			defer wgroup.Done()
			for job := range jobChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if job.db == nil {
					continue
				}

				// 执行数据库查询
				dbInstance, ok := dbIdx[job.db.Identity]
				if !ok {
					errChan <- fmt.Errorf("db instance not exist")
					continue
				}
				query, err := dbInstance.Query(job.inspect.SQL)
				if err != nil {
					errChan <- fmt.Errorf("insp_task get fail: query failed: %s", err)
					continue
				}
				result, err := db2.RowsToResult(query)
				if err != nil {
					errChan <- fmt.Errorf("insp_task rows fail:result conversion failed: %s", err)
					continue
				}

				contentChan <- logger.LogContent{
					Timestamp:   time.Now(),
					TaskName:    taskName.Name,
					TaskID:      taskId,
					InspectName: job.inspect.Identity.Name,
					DBName:      job.db.Identity.Name,
					Result:      result.MarshallJSON(),
				}

				// 发送告警
				if f, err := result.Filter(job.inspect.AlertWhen); err != nil {
					errChan <- fmt.Errorf("insp_task filter fail: %s", err)
				} else if len(f) > 0 {
					if err := alerter2.GetAlert(alertID.Identity()).Send(alerter.Content{
						Success:   true,
						TimeStamp: time.Now(),
						TaskName:  taskName,
						TaskID:    taskId,
						DBName:    job.db.Identity,
						InspName:  job.inspect.Identity,
						Result:    f,
						AlertWhen: job.inspect.AlertWhen,
					}); err != nil {
						errChan <- fmt.Errorf("alert failed: %w", err)
					}
				}
			}
		}()
	}

	// 分发任务
	go func() {
		defer close(jobChan)
		for _, inspect := range tp.inspNodes {
			for _, tdb := range tp.targetDBs {
				select {
				case <-ctx.Done():
					return
				case jobChan <- struct {
					inspect *config.InspConfig
					db      *config.DBConfig
				}{inspect: inspect, db: tdb}:
				}
			}
		}
	}()

	logConents := make([]logger.LogContent, 0, len(dbList)*len(tp.targetDBs))
	go func() {
		for val := range contentChan {
			logConents = append(logConents, val)
		}
	}()

	// 收集错误
	var errors []error
	go func() {
		for err := range errChan {
			errors = append(errors, err)
		}
	}()
	// 等待所有工作完成
	wgroup.Wait()

	//执行后续操作
	close(contentChan)
	err = logger2.Get(t.Config.LogID.Identity()).Log(logConents)
	if err != nil {
		errChan <- err
	}

	close(errChan)
	if len(errors) > 0 {
		// 发送错误告警
		alerter2.GetAlert(alertID.Identity()).Send(alerter.Content{
			TimeStamp: time.Now(),
			Message:   fmt.Sprintf("任务运行出现错误: \n%s", formatErrors(errors)),
			TaskName:  taskName,
			TaskID:    taskId,
		})
		return fmt.Errorf("Async operation error: \n%s", formatErrors(errors))
	}

	fmt.Printf("task submitted: %s\n", t.Config.Identity.Name)
	return nil
}

// 辅助函数：格式化错误信息
func formatErrors(errors []error) string {
	var sb strings.Builder
	for _, err := range errors {
		sb.WriteString(fmt.Sprintf("\n%v\n", err))
	}
	return sb.String()
}

func (t *Task) GetCron() config.CronTab {
	return t.Config.Cron
}

func (t *Task) Identity() config.Identity {
	return t.Config.Identity
}
