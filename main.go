package main

import (
	"WgInspector/app/start"
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"context"
	"log"
	"sync"
)

var (
	global       config.InitConfig
	configPath   = "./app/init_config.yaml"
	mainCtx      context.Context
	mainCancel   context.CancelFunc
	serviceMutex sync.Mutex
)

func main() {

	serviceMutex.Lock()
	defer serviceMutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	mainCtx = ctx
	mainCancel = cancel

	start.Init(global)
	if err := client.Init(global); err != nil {
		log.Printf("[ERROR] Client init: %v", err)
		return
	}

	go func() {
		client.Listen(ctx)
	}()

	log.Println("[INFO] === System services started ===")
}
