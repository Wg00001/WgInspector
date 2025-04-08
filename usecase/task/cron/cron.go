package cron

import (
	"WgInspector/entities/task"
	"context"
	"log"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/3
 */

var (
	globalCron task.Cron
	mu         sync.RWMutex
)

func Use(c task.Cron) {
	mu.Lock()
	defer mu.Unlock()
	if globalCron != nil {
		log.Println("warring: cron register replace")
	}
	if err := c.Init(); err != nil {
		panic(err)
	}
	globalCron = c
	log.Println("cron register and init...")
}

func AddTask(task task.Task) {
	mu.Lock()
	defer mu.Unlock()
	globalCron.AddTask(task)
}

func Start() {
	mu.Lock()
	defer mu.Unlock()
	globalCron.Start()
	log.Println("cron: start...")
}

func Exit() {
	mu.Lock()
	defer mu.Unlock()
	globalCron.Exit()
}

func Monitor(ctx context.Context) (<-chan []task.Stat, error) {
	mu.RLock()
	defer mu.RUnlock()
	return globalCron.Monitor(ctx)
}
