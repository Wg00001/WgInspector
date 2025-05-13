package config

import (
	"WgInspector/utils"
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
	InspNodes  []InspConfig
}

type Identity struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;type:bigserial" json:"ID"`
	Name string `gorm:"unique;notNull" json:"Name"`
}

type IdKey Identity

type DBConfig struct {
	Identity
	Driver string
	DSN    string
}

type LogConfig struct {
	Identity
	Driver string
	Option utils.Option `gorm:"type:jsonb"`
}

type AlertConfig struct {
	Identity
	Driver string
	Option utils.Option `gorm:"type:jsonb"`
}

// ---task(任务)相关配置

type TaskConfig struct {
	Identity
	Cron         CronTab
	AllInspector bool
	LogID        IdKey        `gorm:"type:jsonb"`
	AlertID      IdKey        `gorm:"type:jsonb"`
	TargetDB     []DBConfig   `gorm:"many2many:task_config_target_dbs;"`
	Todo         []InspConfig `gorm:"many2many:task_config_todos;"`
	NotTodo      []InspConfig `gorm:"many2many:task_config_not_todos;"`
}

type CronTab string

//---Agents 相关配置

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
	Cron          CronTab
	LogFilter     LogFilter             `gorm:"type:jsonb"`
	LogID         IdKey                 `gorm:"type:jsonb"`
	AlertID       IdKey                 `gorm:"type:jsonb"`
	AgentID       IdKey                 `gorm:"type:jsonb"`
	KbaseAgentID  IdKey                 `gorm:"type:jsonb"`
	KBase         []KnowledgeBaseConfig `gorm:"many2many:agent_task_config_kbases;"`
	KBaseResults  int
	KBaseMaxLen   int
	SystemMessage string
}

type LogFilter struct {
	// 时间范围：Timestamp 需在 [StartTime, EndTime] 之间
	StartTime time.Time
	EndTime   time.Time
	TaskNames IdKeyArray // Id 匹配列表
	DBIDs     IdKeyArray // DBName 匹配列表
	InspNames IdKeyArray // Insp匹配列表
	TaskIDs   IdKeyArray // TaskID 匹配列表
}

type KnowledgeBaseConfig struct {
	Identity
	Driver  string
	AgentID IdKey                  `gorm:"type:jsonb"`
	Option  map[string]interface{} `gorm:"type:jsonb"`
}
