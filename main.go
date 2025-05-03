package main

import (
	"WgInspector/app/start"
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"context"
	"gopkg.in/yaml.v3"
	"log"
	"os"
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

	file, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(file, &global)
	if err != nil {
		panic(err)
	}

	start.Init(global)
	if err := client.Init(global); err != nil {
		log.Printf("[ERROR] Client init: %v", err)
		return
	}

	go func() {
		start.Run(ctx)
		client.Listen(ctx)
	}()

	log.Println("[INFO] === System services Started ===")

	select {
	case <-ctx.Done():
		log.Println("[INFO] === System services Exited ===")
	}

}
