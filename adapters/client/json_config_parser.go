package client

import (
	config2 "WgInspector/entities/config"
	"WgInspector/usecase/config"
	"encoding/json"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/22
 */

func ParseJsonConfig[T config.ParamType](configType string, configData json.RawMessage) (_ T, err error) {
	switch configType {
	case config2.TypeDB:
		var res config2.DBConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case config2.TypeLog:
		var res config2.LogConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case config2.TypeAlert:
		var res config2.AlertConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case config2.TypeTask:
		var res config2.TaskConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case config2.TypeAgent:
		var res config2.AgentConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case config2.TypeAgentTask:
		var res config2.AgentTaskConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil

	case config2.TypeKBase:
		var res config2.KnowledgeBaseConfig
		if err = json.Unmarshal(configData, &res); err != nil {
			return
		}
		return any(res).(T), nil
	default:
		return
	}
}
