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
	TypeLog       = "Log"
	TypeDB        = "DB"
	TypeAlert     = "Alert"
	TypeTask      = "Task"
	TypeAgent     = "Agent"
	TypeAgentTask = "AgentTask"
	TypeKBase     = "KBase"
	TypeInspector = "Inspector"
)

type ConfigType interface {
	InitConfig | DBConfig | TaskConfig | LogConfig | AlertConfig |
		AgentConfig | AgentTaskConfig | KnowledgeBaseConfig | InspTree | *InspTree
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
	case InspTree, *InspTree:
		return TypeInspector, nil
	default:
		return "", fmt.Errorf("unknown config type: %T", data)
	}
}
