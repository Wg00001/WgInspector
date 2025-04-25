package _default

import (
	"WgInspector/entities/config"
	"WgInspector/entities/logger"
	logger2 "WgInspector/usecase/logger"
	"fmt"
	"log"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/1/19
 */

func init() {
	logger2.RegisterDriver("default", LogDefault{})
}

type LogDefault struct {
}

func (d LogDefault) ReadLog(filter config.LogFilter) ([]logger.LogContent, error) {
	return nil, fmt.Errorf("default logger can't read, pease use other driver")
}

func (d LogDefault) Init(cfg config.LogConfig) (logger.Logger, error) {
	return LogDefault{}, nil
}

func (d LogDefault) GetID() config.Identity {
	return config.Identity{}
}

func (d LogDefault) Log(l logger.LogContent) {
	//utils.PrintQuery(l, rows)
	log.Println(string(l.Result))
}

var _ logger.Logger = (*LogDefault)(nil)
