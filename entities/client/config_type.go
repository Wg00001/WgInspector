package client

import (
	"PgInspector/entities/config"
	"fmt"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/22
 */

const (
	ConfigTypeDB        = "DB"
	ConfigTypeLog       = "Log"
	ConfigTypeAlert     = "Alert"
	ConfigTypeTask      = "Task"
	ConfigTypeAgent     = "Agent"
	ConfigTypeAgentTask = "AgentTask"
	ConfigTypeKBase     = "KBase"
	ConfigTypeInspector = "Inspector"
)

type ConfigType interface {
	config.DefaultConfig | config.DBConfig | config.TaskConfig | config.LogConfig | config.AlertConfig |
		config.AgentConfig | config.AgentTaskConfig | config.KnowledgeBaseConfig | config.InspTree | *config.InspTree
}

func GetConfigTypeName(data any) (string, error) {
	switch data.(type) {
	case config.DBConfig, *config.DBConfig:
		return ConfigTypeDB, nil
	case config.LogConfig, *config.LogConfig:
		return ConfigTypeLog, nil
	case config.AlertConfig, *config.AlertConfig:
		return ConfigTypeAlert, nil
	case config.TaskConfig, *config.TaskConfig:
		return ConfigTypeTask, nil
	case config.AgentConfig, *config.AgentConfig:
		return ConfigTypeAgent, nil
	case config.AgentTaskConfig, *config.AgentTaskConfig:
		return ConfigTypeAgentTask, nil
	case config.KnowledgeBaseConfig, *config.KnowledgeBaseConfig:
		return ConfigTypeKBase, nil
	case config.InspTree, *config.InspTree:
		return ConfigTypeInspector, nil
	default:
		return "", fmt.Errorf("unknown config type: %T", data)
	}
}
