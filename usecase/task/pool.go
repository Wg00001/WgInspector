package task

import (
	"WgInspector/entities/config"
	"context"
	"fmt"
	"sync"
)

/**
 * @description: task pool
 * @author Wg
 * @date 2025/2/11
 */

var pool = sync.Map{}

func Register(t *Task) error {
	if t == nil {
		return fmt.Errorf("task init err, build task fail")
	}
	pool.Store(t.Config.Identity, t)
	return nil
}

func Get(name config.Identity) *Task {
	if val, ok := pool.Load(name); ok {
		return val.(*Task)
	}
	return nil
}

func Delete(name config.Identity) {
	pool.Delete(name)
}

func Do(name config.Identity) error {
	t := Get(name)
	if t == nil {
		return fmt.Errorf("name of task not exist, name: %s\n", name)
	}
	return t.Do(context.Background())
}
