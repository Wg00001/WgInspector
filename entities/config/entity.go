package config

import (
	"WgInspector/utils"
	"encoding/json"
	"time"
)

/**
 * @description: 配置的实体定义
 * @author Wg
 * @date 2025/1/19
 */

type InitConfig struct {
	BaseDSN   string `yaml:"base_dsn"`
	ClientURL string `yaml:"client_url"`
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
	ID   int64  `gorm:"primaryKey;autoIncrement;type:bigserial" json:"ID"`
	Name string `json:"Name"`
}

type DBConfig struct {
	Identity
	Driver string
	DSN    string
}

type LogConfig struct {
	Identity
	Driver string
	Option json.RawMessage `gorm:"type:jsonb"`
}

type AlertConfig struct {
	Identity
	Driver string
	Option json.RawMessage `gorm:"type:jsonb"`
}

// ---task(任务)相关配置

type TaskConfig struct {
	Identity
	Cron         Cron `gorm:"type:jsonb"`
	AllInspector bool

	LogID    Identity   `gorm:"type:jsonb"`
	AlertID  Identity   `gorm:"type:jsonb"`
	TargetDB []Identity `gorm:"type:jsonb"`

	Todo    []Identity `gorm:"type:jsonb"`
	NotTodo []Identity `gorm:"type:jsonb"`
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
	Cron          Cron       `gorm:"type:jsonb"`
	LogFilter     LogFilter  `gorm:"type:jsonb"`
	LogID         Identity   `gorm:"type:jsonb"`
	AlertID       Identity   `gorm:"type:jsonb"`
	AgentID       Identity   `gorm:"type:jsonb"`
	KbaseAgentID  Identity   `gorm:"type:jsonb"`
	KBase         []Identity `gorm:"type:jsonb"`
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
	Driver  string
	AgentID Identity               `gorm:"type:jsonb"`
	Option  map[string]interface{} `gorm:"type:jsonb"`
}
