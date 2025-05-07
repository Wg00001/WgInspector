package client

import (
	"WgInspector/entities/client"
	"WgInspector/entities/config"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/25
 */

const defaultClientDB = "postgres"

func Init(cfg config.InitConfig) error {
	driversMu.Lock()
	defer driversMu.Unlock()
	init, err := aclient.Init(cfg.ClientURL)
	if err != nil {
		return err
	}
	err = registerUseClientDatabase(cfg.Option.GetOrDefault("client_db_driver", defaultClientDB))
	if err != nil {
		return err
	}
	return Register(init)
}

var (
	aclient   client.Client
	driversMu sync.Mutex
)

func RegisterClient(cli client.Client) {
	driversMu.Lock()
	defer driversMu.Unlock()
	aclient = cli
}
