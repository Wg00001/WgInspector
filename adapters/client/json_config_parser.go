package client

import (
	"PgInspector/entities/client"
	config2 "PgInspector/entities/config"
	"PgInspector/usecase/config"
	"encoding/json"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/22
 */

func ParseJsonConfig2[T config.ParamType](configType string, configData json.RawMessage) (_ T) {
	switch configType {
	case client.ConfigTypeDB:
		var res config2.DBConfig
		if err := json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(config2.DBConfig)
	}
}

func ParseJsonConfig[T config.ParamType](configType string, configData json.RawMessage) (_ T, err error) {
	switch configType {
	case client.ConfigTypeDB:
		var res config2.DBConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case client.ConfigTypeLog:
		var res config2.LogConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case client.ConfigTypeAlert:
		var res config2.AlertConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case client.ConfigTypeTask:
		var res config2.TaskConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case client.ConfigTypeAgent:
		var res config2.AgentConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case client.ConfigTypeAgentTask:
		var res config2.AgentTaskConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case client.ConfigTypeKBase:
		var res config2.KnowledgeBaseConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil
	default:
		return
	}
}
