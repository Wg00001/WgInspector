package client

import (
	"WgInspector/entities/client"
	"WgInspector/entities/config"
	"fmt"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/25
 */

func Use(cfg config.DefaultConfig) error {
	driver, err := GetDriver(cfg.ClientDriver)
	if err != nil {
		return err
	}
	init, err := driver.Init(cfg.ClientURL)
	if err != nil {
		return err
	}
	return Register(init)
}

var (
	drivers   = make(map[string]client.Client)
	driversMu sync.Mutex
)

func RegisterDriver(name string, cli client.Client) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if drivers == nil {
		panic("client: drivers map is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("client: Register called twice for driver " + name)
	}
	drivers[name] = cli
}

func GetDriver(name string) (client.Client, error) {
	driversMu.Lock()
	defer driversMu.Unlock()
	res, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("client: get driver fail - %s\n", name)
	}
	return res, nil
}
