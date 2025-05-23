package config

import (
	"WgInspector/entities/config"
	"WgInspector/utils"
	"fmt"
	"github.com/wg00001/wgo-sdk/wg"
	"sync"
)

/**
 * @description: 配置中心
 * @author Wg
 * @date 2025/2/4
 */

var (
	// Meta : 对pgsql中数据的缓存
	Meta      = config.MetaConfig{}
	index     = make(map[Key]*config.Id) //存储配置文件的指针，get时返回的是解引用后的实体
	inspIndex *config.InspIndex
	mu        sync.RWMutex
	appendMU  sync.Mutex

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

func GetAllInsp() []config.InspConfig {
	mu.RLock()
	defer mu.RUnlock()
	return Meta.InspNodes
}

func GetInsp(id config.Identity, ids ...config.Identity) []*config.InspConfig {
	mu.RLock()
	defer mu.RUnlock()
	return inspIndex.Get(id, ids...)
}

func GetInspRoot() []config.InspConfig {
	mu.RLock()
	defer mu.RUnlock()
	return inspIndex.GetForestRootList()
}

func SetInitConfig(initConfig config.InitConfig) {
	globalInitConfig = initConfig
}

func GetInitConfig() config.InitConfig {
	return globalInitConfig
}

func GetMetaItem(configType string) (any, error) {
	mu.RLock()
	defer mu.RUnlock()
	switch configType {
	case config.TypeDB:
		return sliceCopy(Meta.DBs), nil
	case config.TypeLog:
		return sliceCopy(Meta.Logs), nil
	case config.TypeAlert:
		return sliceCopy(Meta.Alerts), nil
	case config.TypeTask:
		return sliceCopy(Meta.Tasks), nil
	case config.TypeAgent:
		return sliceCopy(Meta.Agents), nil
	case config.TypeAgentTask:
		return sliceCopy(Meta.AgentTasks), nil
	case config.TypeKBase:
		return sliceCopy(Meta.KBases), nil
	case config.TypeInspector:
		return GetInspRoot(), nil
	default:
		return nil, fmt.Errorf("config-center: get meta item fail, type: %s", configType)
	}
}

func GetIdentityList(configType string) ([]config.Identity, error) {
	mu.RLock()
	defer mu.RUnlock()
	switch configType {
	case config.TypeDB:
		return wg.SliceToSlice(Meta.DBs, func(item config.DBConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeLog:
		return wg.SliceToSlice(Meta.Logs, func(item config.LogConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeAlert:
		return wg.SliceToSlice(Meta.Alerts, func(item config.AlertConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeTask:
		return wg.SliceToSlice(Meta.Tasks, func(item config.TaskConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeAgent:
		return wg.SliceToSlice(Meta.Agents, func(item config.AgentConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeAgentTask:
		return wg.SliceToSlice(Meta.AgentTasks, func(item config.AgentTaskConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeKBase:
		return wg.SliceToSlice(Meta.KBases, func(item config.KnowledgeBaseConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	case config.TypeInspector:
		return wg.SliceToSlice(Meta.InspNodes, func(item config.InspConfig) config.Identity {
			return item.GetIdentity()
		}), nil
	default:
		return nil, fmt.Errorf("config-center: get meta item fail, type: %s", configType)
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
	inspIndex = config.NewInspIndex(Meta.InspNodes)
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
		inspIndex = config.NewInspIndex(Meta.InspNodes)
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

func updateConfigRoots(val config.Id) error {
	v, ok := val.(config.InspConfig)
	if !ok {
		inspIndex = config.NewInspIndex(Meta.InspNodes)
	}
	return inspIndex.Put(v)
}
