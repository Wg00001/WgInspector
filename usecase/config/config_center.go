package config

import (
	"WgInspector/entities/config"
	"WgInspector/utils"
	"fmt"
	"sync"
)

/**
 * @description: 配置中心
 * @author Wg
 * @date 2025/2/4
 */

var (
	// Meta : 对pgsql中数据的缓存
	Meta     = config.MetaConfig{}
	index    = make(map[Key]*config.Id)
	mu       sync.RWMutex
	appendMU sync.Mutex

	globalInitConfig config.InitConfig
)

type Key struct {
	ConfigType string
	config.Identity
}

func RLock() {
	mu.RLock()
}

func RUnlock() {
	mu.RUnlock()
}

//func GetInsp(path config.Identity) *config.InspConfig {
//	mu.RLock()
//	defer mu.RUnlock()
//	return Meta.InspIndex.GetNode(path.ToString())
//}

func GetAllInsp() []config.InspConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Meta.InspNodes
}

func GetInsp(id config.Identity, ids ...config.Identity) []*config.InspConfig {
	if Meta.InspIndex == nil {
		Meta.InspIndex = config.NewInspIndex(Meta.InspNodes)
	}
	return Meta.InspIndex.Get(id, ids...)
}

func SetInitConfig(initConfig config.InitConfig) {
	globalInitConfig = initConfig
}

func GetInitConfig() config.InitConfig {
	return globalInitConfig
}

// 可选的入参T，仅用于指定T的类型
func GetMetaItem[T config.Id](...T) ([]T, error) {
	mu.RLock()
	defer mu.RUnlock()
	var t T
	switch any(t).(type) {
	case config.DBConfig:
		return sliceCopy(Meta.DBs).([]T), nil
	case config.LogConfig:
		return sliceCopy(Meta.Logs).([]T), nil
	case config.AlertConfig:
		return sliceCopy(Meta.Alerts).([]T), nil
	case config.TaskConfig:
		return sliceCopy(Meta.Tasks).([]T), nil
	case config.AgentConfig:
		return sliceCopy(Meta.Agents).([]T), nil
	case config.AgentTaskConfig:
		return sliceCopy(Meta.AgentTasks).([]T), nil
	case config.KnowledgeBaseConfig:
		return sliceCopy(Meta.KBases).([]T), nil
	case config.InspConfig:
		return sliceCopy(Meta.InspNodes).([]T), nil
	default:
		return nil, fmt.Errorf("unknown config type: %T", t)
	}
}

func GetConfigMeta() config.MetaConfig {
	mu.RLock()
	defer mu.RUnlock()
	var res config.MetaConfig
	utils.DeepCopy(res, Meta)
	return res
}

func SetConfigMeta(c config.MetaConfig) error {
	mu.Lock()
	Meta = c
	mu.Unlock()
	AppendIndex(config.TypeAlert, c.Alerts...)
	AppendIndex(config.TypeDB, c.DBs...)
	AppendIndex(config.TypeLog, c.Logs...)
	AppendIndex(config.TypeTask, c.Tasks...)
	AppendIndex(config.TypeAgent, c.Agents...)
	AppendIndex(config.TypeAgentTask, c.AgentTasks...)
	AppendIndex(config.TypeKBase, c.KBases...)
	AppendIndex(config.TypeInspector, c.InspNodes...)
	return nil
}

func AppendIndex[T config.Id](configTypes string, configs ...T) {
	appendMU.Lock()
	defer appendMU.Unlock()
	key := Key{ConfigType: configTypes}
	for i := range configs {
		key.Identity = configs[i].GetIdentity()
		a := any(configs[i]).(config.Id)
		index[key] = &a
	}
	return
}

func Get(key Key) (res config.Id, err error) {
	mu.RLock()
	defer mu.RUnlock()
	data, ok := index[key]
	if !ok {
		err = fmt.Errorf("config_center: entry not exist - key:[%v]", key)
		return
	}
	res, ok = any(*data).(config.Id)
	if !ok {
		err = fmt.Errorf("config_center: type of entry not allow - key:[%v]", key)
		return
	}
	return
}

func GetWithType[T config.ConfigType](key Key) (res T, err error) {
	mu.RLock()
	defer mu.RUnlock()
	data, ok := index[key]
	if !ok {
		err = fmt.Errorf("config_center: entry not exist - key:[%v]", key)
		return
	}
	res, ok = any(*data).(T)
	if !ok {
		err = fmt.Errorf("config_center: type of entry not allow - key:[%v]", key)
		return
	}
	return
}

func Save(key Key, val config.Id) (err error) {
	mu.Lock()
	defer mu.Unlock()
	key.ID, err = reader.SaveConfig(val)
	if err != nil {
		return err
	}
	//_, ok := index[key]
	//if !ok {
	//	AppendIndex(key.ConfigType, val)
	//} else {
	//	return util.DeepCopy(index[key], &val)
	//}
	return syncWithDB(key.ConfigType)
}

func Del(key Key, val config.Id) error {
	mu.Lock()
	defer mu.Unlock()
	err := reader.DeleteConfig(val)
	if err != nil {
		return err
	}
	delete(index, key)
	return syncWithDB(key.ConfigType)
}

func syncWithDB(configType string) error {
	meta, err := reader.ReadConfig(configType)
	if err != nil {
		return err
	}
	switch configType {
	case config.TypeDB:
		Meta.DBs = meta.DBs
		AppendIndex(config.TypeDB, Meta.DBs...)
	case config.TypeLog:
		Meta.Logs = meta.Logs
		AppendIndex(config.TypeLog, Meta.Logs...)
	case config.TypeAlert:
		Meta.Alerts = meta.Alerts
		AppendIndex(config.TypeAlert, Meta.Alerts...)
	case config.TypeTask:
		Meta.Tasks = meta.Tasks
		AppendIndex(config.TypeTask, Meta.Tasks...)
	case config.TypeAgent:
		Meta.Agents = meta.Agents
		AppendIndex(config.TypeAgent, Meta.Agents...)
	case config.TypeAgentTask:
		Meta.AgentTasks = meta.AgentTasks
		AppendIndex(config.TypeAgentTask, Meta.AgentTasks...)
	case config.TypeKBase:
		Meta.KBases = meta.KBases
		AppendIndex(config.TypeKBase, Meta.KBases...)
	case config.TypeInspector:
		Meta.InspNodes = meta.InspNodes
		AppendIndex(config.TypeInspector, Meta.InspNodes...)
	default:
		return fmt.Errorf("unsupported config type: %s", configType)
	}
	return nil
}

func sliceCopy[T config.Id](arr []T) any {
	res := make([]T, len(arr))
	copy(res, arr)
	return res
}
