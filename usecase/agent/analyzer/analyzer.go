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
 * @date 2025/2/25
 */

var (
	pool = sync.Map{}
)

func Register(aiConfig config.AgentConfig) error {
	driver, err := GetDriver(aiConfig.Driver)
	if err != nil {
		return err
	}
	init, err := driver.Init(&aiConfig)
	if err != nil {
		return err
	}
	pool.Store(aiConfig.Identity, init)
	return nil
}

//根据driver找到对应的adapter实现，以init全局analyzer

func Get(agentID config.Identity) (agent.Analyzer, error) {
	a, ok := pool.Load(agentID)
	if !ok {
		return nil, fmt.Errorf("analyzer: config not exist: %s\n", agentID)
	}
	res, ok := a.(agent.Analyzer)
	if !ok {
		return nil, fmt.Errorf("analyzer: type err: %s\n", agentID)
	}
	return res, nil
}
