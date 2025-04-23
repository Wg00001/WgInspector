package config

import (
	"WgInspector/entities/config"
	"fmt"
	"sync"
)

/**
 * @description: 配置中心
 * @author Wg
 * @date 2025/2/4
 */

var (
	// Meta 对pgsql中数据的缓存
	Meta     = config.MetaConfig{Insp: config.NewTree()}
	index    = make(map[Key]config.Id)
	mu       sync.RWMutex
	appendMU sync.Mutex
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

func GetInsp(path config.Identity) *config.InspNode {
	mu.RLock()
	defer mu.RUnlock()
	return Meta.Insp.GetNode(path.ToString())
}

func GetAllInsp() []*config.InspNode {
	mu.RLock()
	defer mu.RUnlock()
	return Meta.Insp.AllInsp
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
		index[key] = configs[i]
	}
	return
}

func Get[T config.Id](key Key) (res T, err error) {
	mu.RLock()
	defer mu.RUnlock()
	data, ok := index[key]
	if !ok {
		err = fmt.Errorf("config_center: entry not exist - key:[%v]", key)
		return
	}
	res, ok = any(data).(T)
	if !ok {
		err = fmt.Errorf("config_center: type of entry not allow - key:[%v]", key)
		return
	}
	return
}

func Save(key Key, val config.Id) (err error) {
	mu.RLock()
	defer mu.RUnlock()
	err = reader.SaveConfig(val)
	if err != nil {
		return err
	}
	delete(index, key)
	AppendIndex(key.ConfigType, val)
	return nil
}

func Del(key Key, val config.Id) error {
	mu.RLock()
	defer mu.RUnlock()
	err := reader.DeleteConfig(val)
	if err != nil {
		return err
	}
	delete(index, key)
	AppendIndex(key.ConfigType, val)
	return nil
}
