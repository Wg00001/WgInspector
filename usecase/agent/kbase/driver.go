package kbase

import (
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	"fmt"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/5
 */

func Use(kbaseConfig config.KnowledgeBaseConfig) error {
	driver, err := GetDriver(kbaseConfig.Driver)
	if err != nil {
		return err
	}
	init, err := driver.Init(&kbaseConfig)
	if err != nil {
		return err
	}
	return Register(kbaseConfig.Identity, init)
}

var (
	drivers  = make(map[string]agent.KnowledgeBase)
	muDriver sync.RWMutex
)

func RegisterDriver(name string, driver agent.KnowledgeBase) {
	muDriver.Lock()
	defer muDriver.Unlock()
	if drivers == nil {
		panic("agent: kbase drivers map is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("agent: kbase Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func GetDriver(name string) (agent.KnowledgeBase, error) {
	muDriver.RLock()
	defer muDriver.RUnlock()
	res, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("agent: get kbase driver fail - %s\n", name)
	}
	return res, nil
}
