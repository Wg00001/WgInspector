package json

import (
	"WgInspector/entities/config"
	"encoding/json"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/18
 */

type ConfigJsonParser struct {
	config.ConfigMeta
}

var _ config.Parser = (*ConfigJsonParser)(nil)

func (c ConfigJsonParser) ParseConfig(bytes []byte) (res config.CommonConfigGroup, err error) {
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		res.DBs = func() (arr []config.DBConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return nil
			}
			return arr
		}()
		if err != nil {
			return
		}

		res.Logs = func() (arr []config.LogConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return nil
			}
			return arr
		}()
		if err != nil {
			return
		}

		res.Alerts = func() (arr []config.AlertConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return nil
			}
			return arr
		}()
	}
	return
}

func (c ConfigJsonParser) ParseTask(bytes []byte) (config.TaskConfigGroup, error) {
	var res config.TaskConfigGroup
	err := json.Unmarshal(bytes, &res)
	if err != nil {
		res.Tasks = func() (arr []config.TaskConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return nil
			}
			return arr
		}()
	}
	return res, err
}

func (c ConfigJsonParser) ParseInspector(bytes []byte) (res *config.InspTree, err error) {
	err = json.Unmarshal(bytes, &res)
	return
}

func (c ConfigJsonParser) ParseAgent(bytes []byte) (res config.AgentConfigGroup, err error) {
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		res.Agent = func() (arr config.AgentConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return config.AgentConfig{}
			}
			return arr
		}()
		if err != nil {
			return
		}

		res.AgentTasks = func() (arr []config.AgentTaskConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return nil
			}
			return arr
		}()
		if err != nil {
			return
		}

		res.KnowledgeBases = func() (arr []config.KnowledgeBaseConfig) {
			err = json.Unmarshal(bytes, &arr)
			if err != nil {
				return nil
			}
			return arr
		}()
	}
	return
}

func (c ConfigJsonParser) Encode(arg any) ([]byte, error) {
	return json.Marshal(arg)
}
