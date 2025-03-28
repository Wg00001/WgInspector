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
	DefaultConfig | DBConfig | TaskConfig | LogConfig | AlertConfig |
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

func GetConfigType[T ConfigType](typeName string) (T, error) {
	switch typeName {
	case TypeLog:
		return any(LogConfig{}).(T), nil
	case TypeDB:
		return any(DBConfig{}).(T), nil
	case TypeAlert:
		return any(AlertConfig{}).(T), nil
	case TypeTask:
		return any(TaskConfig{}).(T), nil
	case TypeAgent:
		return any(AgentConfig{}).(T), nil
	case TypeAgentTask:
		return any(AgentTaskConfig{}).(T), nil
	case TypeKBase:
		return any(KnowledgeBaseConfig{}).(T), nil
	case TypeInspector:
		return any(InspTree{}).(T), nil
	default:
		return any(LogConfig{}).(T), fmt.Errorf("unknown config type: " + typeName)
	}
}
