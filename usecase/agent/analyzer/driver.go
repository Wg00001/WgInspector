package analyzer

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

func Use(aiConfig config.AgentConfig) error {
	driver, err := GetDriver(aiConfig.Driver)
	if err != nil {
		return err
	}
	init, err := driver.Init(&aiConfig)
	if err != nil {
		return err
	}
	Register(init)
	return nil
}

var (
	drivers  = make(map[string]agent.Analyzer)
	muDriver sync.RWMutex
)

func RegisterDriver(name string, driver agent.Analyzer) {
	muDriver.Lock()
	defer muDriver.Unlock()
	if drivers == nil {
		panic("agent: drivers map is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("agent: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func GetDriver(name string) (agent.Analyzer, error) {
	muDriver.RLock()
	defer muDriver.RUnlock()
	res, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("agent: get driver fail - %s\n", name)
	}
	return res, nil
}
