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
 * @date 2025/2/15
 */

var pool = sync.Map{}

func Register(lg logger.Logger) error {
	if _, ok := pool.Load(lg.GetID()); ok {
		return fmt.Errorf("logger register fail: logger is already exsit, loggerID repeat")
	}
	pool.Store(lg.GetID(), lg)
	return nil
}

func Get(id config.Identity) logger.Logger {
	val, ok := pool.Load(id)
	if !ok {
		res, _ := GetDriver("default")
		return res
	}
	t, ok := val.(logger.Logger)
	if !ok {
		res, _ := GetDriver("default")
		return res
	}
	return t
}
