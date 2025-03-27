package config

import (
	"WgInspector/utils"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/4
 */

type Reader interface {
	NewReader(option utils.Option) (Reader, error)
	ReadConfig() error
	SaveIntoConfig()
	Watch() //todo
	//todo Save()
}

type Parser interface {
	ParseConfig([]byte) (CommonConfigGroup, error)
	ParseTask([]byte) (TaskConfigGroup, error)
	ParseInspector([]byte) (*InspTree, error)
	ParseAgent([]byte) (AgentConfigGroup, error)
}
