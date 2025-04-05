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

var pool = sync.Map{}

func Register(name config.Identity, lg agent.KnowledgeBase) error {
	if _, ok := pool.Load(name); ok {
		return fmt.Errorf("agent kbase register fail: kbase is already exsit, name repeat - %s\n", name)
	}
	pool.Store(name, lg)
	return nil
}

func Get(name config.Identity) agent.KnowledgeBase {
	val, ok := pool.Load(name)
	if !ok {
		res, _ := GetDriver("default")
		return res
	}
	t, ok := val.(agent.KnowledgeBase)
	if !ok {
		res, _ := GetDriver("default")
		return res
	}
	return t
}

func Save(docs []agent.Document) (err error) {
	pool.Range(func(key, value any) bool {
		val, ok := value.(agent.KnowledgeBase)
		if !ok {
			err = fmt.Errorf("type turn fail")
			return false
		}
		err = val.WriteIn(docs)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		err = fmt.Errorf("kbase save fail: %s", err)
	}
	return
}
