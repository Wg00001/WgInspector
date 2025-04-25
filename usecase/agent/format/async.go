package format

import (
	"WgInspector/entities/logger"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/24
 */

type TaskGroups struct {
	Tasks        map[string]*Task
	isAsync      bool
	sync.RWMutex                     // 全局读写锁
	taskChan     chan asyncTaskEvent // 异步任务事件通道
}

type asyncTaskEvent struct {
	taskName string
	content  *logger.LogContent
}

type Task struct {
	TaskName string         `json:"task_name"`
	DBGroups map[string]*DB `json:"db_groups"`

	sync.Mutex `json:"-"`
	queue      chan *logger.LogContent `json:"-"` // 每个 Task 独立的消息队列
}

// 初始化 TaskGroups 时启动调度器
func NewTaskGroups(bufferSize int) *TaskGroups {
	tg := &TaskGroups{
		Tasks:    make(map[string]*Task),
		taskChan: make(chan asyncTaskEvent, bufferSize), // 全局任务分发通道
	}
	return tg
}

func (tg *TaskGroups) Async() {
	tg.isAsync = true
	go tg.scheduler() // 启动全局调度协程
}

// 全局调度器（负责分配任务到各 Task 队列）
func (tg *TaskGroups) scheduler() {
	for event := range tg.taskChan {
		tg.RLock()
		task, exists := tg.Tasks[event.taskName]
		tg.RUnlock()

		if !exists {
			// 双重检查锁创建新 Task
			tg.Lock()
			task, exists = tg.Tasks[event.taskName]
			if !exists {
				task = &Task{
					TaskName: event.taskName,
					DBGroups: make(map[string]*DB),
					queue:    make(chan *logger.LogContent, 100), // 每个 Task 独立队列
				}
				tg.Tasks[event.taskName] = task
				go task.processQueue() // 启动 Task 专属消费协程
			}
			tg.Unlock()
		}

		// 将内容投递到对应 Task 队列
		task.queue <- event.content
	}
}

// Task 专属消费协程
func (t *Task) processQueue() {
	for content := range t.queue {
		t.Append(content) // 同步追加逻辑
	}
}
