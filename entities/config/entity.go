package config

import (
	"WgInspector/utils"
	"github.com/google/uuid"
	"time"
)

/**
 * @description: 配置的实体定义
 * @author Wg
 * @date 2025/1/19
 */

type InitConfig struct {
	BaseDSN   string
	ClientURL string
	Option    utils.Option
}

type MetaConfig struct {
	DBs        []DBConfig
	Logs       []LogConfig
	Alerts     []AlertConfig
	Tasks      []TaskConfig
	Agents     []AgentConfig
	AgentTasks []AgentTaskConfig
	KBases     []KnowledgeBaseConfig
	InspNodes  []InspNode
	Insp       *InspTree
}

type Identity struct {
	UUID uuid.UUID
	Name string
}

type DBConfig struct {
	Identity
	Driver string
	DSN    string
}

type LogConfig struct {
	Identity
	Driver string
	Option map[string]string
}

type AlertConfig struct {
	Identity
	Driver string
	Option map[string]string
}

// ---task(任务)相关配置

type TaskConfig struct {
	Identity
	Cron         *Cron
	AllInspector bool

	LogID    Identity
	AlertID  Identity
	TargetDB []Identity

	Todo    []Identity
	NotTodo []Identity
}

type Cron struct {
	CronTab  string
	Duration time.Duration
	AtTime   []string
	Weekly   []time.Weekday
	Monthly  []int
}

// ---Agents Agents 相关配置

// AgentConfig 用户只能指定一个全局Ai，所有的分析均由此Ai完成
type AgentConfig struct {
	Identity
	Driver        string
	Url           string
	ApiKey        string
	Model         string
	Temperature   float64
	SystemMessage string
}

type AgentTaskConfig struct {
	Identity
	Cron          *Cron
	LogID         Identity
	LogFilter     LogFilter
	AlertID       Identity
	KBase         []Identity
	KBaseResults  int
	KBaseMaxLen   int
	SystemMessage string
}

type LogFilter struct {
	// 时间范围：Timestamp 需在 [StartTime, EndTime] 之间
	StartTime time.Time
	EndTime   time.Time
	TaskNames []Identity // Id 匹配列表
	DBIDs     []Identity // DBName 匹配列表
	TaskIDs   []Identity // TaskID 匹配列表
	InspNames []Identity // Insp匹配列表
}

type KnowledgeBaseConfig struct {
	Identity
	Driver string
	Value  map[string]interface{}
}
