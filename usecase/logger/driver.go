package logger

import (
	"WgInspector/entities/config"
	"WgInspector/entities/logger"
	"fmt"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/4
 */

func Use(cfg config.LogConfig) error {
	driver, err := GetDriver(cfg.Driver)
	if err != nil {
		return err
	}
	init, err := driver.Init(cfg)
	if err != nil {
		return err
	}
	return Register(init)
}

var (
	drivers = make(map[string]logger.Logger)
	mu      sync.RWMutex
)

func RegisterDriver(name string, driver logger.Logger) {
	mu.Lock()
	defer mu.Unlock()
	if drivers == nil {
		panic("logger: drivers map is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("logger: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func GetDriver(name string) (logger.Logger, error) {
	mu.RLock()
	defer mu.RUnlock()
	res, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("logger: get driver fail %s\n", name)
	}
	return res, nil
}
