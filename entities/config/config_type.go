package config

import (
	"fmt"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/22
 */

const (
	TypeLog       = "log_config"
	TypeDB        = "db_config"
	TypeAlert     = "alert_config"
	TypeTask      = "task_config"
	TypeAgent     = "agent_config"
	TypeAgentTask = "agent_task_config"
	TypeKBase     = "kbase_config"
	TypeInspector = "inspector_config"
)

type ConfigType interface {
	DBConfig | TaskConfig | LogConfig | AlertConfig |
		AgentConfig | AgentTaskConfig | KnowledgeBaseConfig | InspConfig
}

func Turn[T ConfigType](data Id) T {
	return any(data).(T)
}

func GetConfigTypeName(data any) (string, error) {
	switch data.(type) {
	case DBConfig, *DBConfig:
		return TypeDB, nil
	case LogConfig, *LogConfig:
		return TypeLog, nil
	case AlertConfig, *AlertConfig:
		return TypeAlert, nil
	case TaskConfig, *TaskConfig:
		return TypeTask, nil
	case AgentConfig, *AgentConfig:
		return TypeAgent, nil
	case AgentTaskConfig, *AgentTaskConfig:
		return TypeAgentTask, nil
	case KnowledgeBaseConfig, *KnowledgeBaseConfig:
		return TypeKBase, nil
	case InspConfig, *InspConfig:
		return TypeInspector, nil
	default:
		return "", fmt.Errorf("config_type: unknown config type: %T", data)
	}
}

func (t DBConfig) TableName() string {
	return TypeDB
}

func (l LogConfig) TableName() string {
	return TypeLog
}

func (a AlertConfig) TableName() string {
	return TypeAlert
}

func (t TaskConfig) TableName() string {
	return TypeTask
}

func (a AgentConfig) TableName() string {
	return TypeAgent
}

func (a AgentTaskConfig) TableName() string {
	return TypeAgentTask
}

func (k KnowledgeBaseConfig) TableName() string {
	return TypeKBase
}

func (i InspConfig) TableName() string {
	return TypeInspector
}
