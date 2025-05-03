package format

import (
	"WgInspector/entities/logger"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

/**
 * @description: 用于将LogContent转成可供AI读取的string格式
 * @author Wg
 * @date 2025/2/24
 */

// Format 分类格式化
func Format(contents ...logger.LogContent) (*string, error) {
	taskGroups := NewTaskGroups(100)
	for _, c := range contents {
		taskGroups.SyncAppend(&c)
	}

	for _, task := range taskGroups.Tasks {
		for _, db := range task.DBGroups {
			for _, insp := range db.InspGroups {
				sort.Slice(insp.Contents, func(i, j int) bool { //升序排序
					// 降序改为 tj.Before(ti)
					return insp.Contents[i].Timestamp.Before(insp.Contents[j].Timestamp)
				})
			}
		}
	}

	// 生成 JSON
	var sb strings.Builder
	encoder := json.NewEncoder(&sb)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(taskGroups); err != nil {
		return nil, fmt.Errorf("AI send err: encoding fail: %v", err)
	}
	res := sb.String()
	return &res, nil
}

type DB struct {
	DBName     string              `json:"group_name"`
	InspGroups map[string]*Inspect `json:"insp_groups"`
}

type Inspect struct {
	InspName string     `json:"insp_name"`
	Contents []*Content `json:"inspects"`
}

type Content struct {
	Timestamp time.Time `json:"timestamp"`
	Result    string    `json:"result"` // 结果（JSON字符串或原始文本）
}

// SyncAppend 将格式化的过程异步进行
func (tg *TaskGroups) SyncAppend(c *logger.LogContent) {
	tg.Lock()
	defer tg.Unlock()

	if _, ok := tg.Tasks[c.TaskName]; !ok {
		tg.Tasks[c.TaskName] = &Task{
			TaskName: c.TaskName,
			DBGroups: make(map[string]*DB),
		}
	}
	tg.Tasks[c.TaskName].Append(c)
}

func (tg *TaskGroups) AsyncAppend(c *logger.LogContent) {
	if !tg.isAsync {
		tg.SyncAppend(c)
		log.Println("AI send warring: This group has not turned on asynchronous and has automatically switched to synchronous execution")
		return
	}
	tg.taskChan <- asyncTaskEvent{
		taskName: c.TaskName,
		content:  c,
	}
}

func (t *Task) Append(c *logger.LogContent) {
	if _, ok := t.DBGroups[c.DBName]; !ok {
		t.DBGroups[c.DBName] = &DB{
			DBName:     c.DBName,
			InspGroups: make(map[string]*Inspect),
		}
	}
	db := t.DBGroups[c.DBName]
	if _, ok := db.InspGroups[c.InspectName]; !ok {
		db.InspGroups[c.InspectName] = &Inspect{
			InspName: c.InspectName,
			Contents: []*Content{},
		}
	}
	insp := db.InspGroups[c.InspectName]
	insp.Contents = append(insp.Contents, convertToAIContent(c))
}

// 辅助函数：转换单个 Content
func convertToAIContent(content *logger.LogContent) *Content {
	return &Content{
		Timestamp: content.Timestamp,
		Result:    string(content.Result),
	}
}
