package config

import (
	"time"
)

/**
 * @description: 配置的实体定义
 * @author Wg
 * @date 2025/1/19
 */

type DefaultConfig struct {
	ConfigReader string
	ConfigParser string
	ClientDriver string
	ClientURL    string
}

type ConfigMeta struct {
	CommonConfigGroup
	TaskConfigGroup
	AgentConfigGroup
	Insp *InspTree
}

type CommonConfigGroup struct {
	DBs    []DBConfig
	Logs   []LogConfig
	Alerts []AlertConfig
}

// TaskConfigGroup 可以被Agent修改的
type TaskConfigGroup struct {
	Tasks []TaskConfig
	//Insp  insp.InspTree
}

type AgentConfigGroup struct {
	Agent          AgentConfig
	AgentTasks     []AgentTaskConfig
	KnowledgeBases []KnowledgeBaseConfig
}

type ConfigIndex struct {
	Default *DefaultConfig
	Task    map[Identity]*TaskConfig
	DB      map[Identity]*DBConfig
	Log     map[Identity]*LogConfig
	Alert   map[Identity]*AlertConfig

	Agent     *AgentConfig
	AgentTask map[Identity]*AgentTaskConfig
	KBase     map[Identity]*KnowledgeBaseConfig
	//Insp      *insp.InspTree
}

type Identity string

type DBConfig struct {
	Identity
	Driver string
	DSN    string
}

type LogConfig struct {
	Identity
	Driver string
	Header map[string]string
}

type AlertConfig struct {
	Identity
	Driver string
	Header map[string]string
}

// ---task(任务)相关配置

type TaskConfig struct {
	Identity
	Cron         *Cron
	AllInspector bool

	LogID    Identity
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

// ---Agent Agent 相关配置

// AgentConfig 用户只能指定一个全局Ai，所有的分析均由此Ai完成
type AgentConfig struct {
	//AiName      Id
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
	DBNames   []Identity // DBName 匹配列表
	TaskIDs   []string   // TaskID 匹配列表
	InspNames []Identity // Insp匹配列表
}

type KnowledgeBaseConfig struct {
	Identity
	Driver string
	Value  map[string]interface{}
}
