package main

import (
	"WgInspector/app/start"
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"context"
	"gopkg.in/yaml.v3"

	// "gopkg.in/yaml.v3"
	"log"
	_ "net/http/pprof"
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

	//go func() {
	//	log.Printf("性能监控服务器运行在 http://localhost:%d/debug/pprof\n", 6060)
	//	log.Fatal(http.ListenAndServe(":6060", nil))
	//}()

	global = config.InitConfig{
		BaseDSN:   os.Getenv("BASE_DSN"),
		ClientURL: os.Getenv("CLIENT_URL"),
		Option:    map[string]string{},
	}

	if global.ClientURL == "" {
		file, err := os.ReadFile(configPath)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(file, &global)
		if err != nil {
			panic(err)
		}
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

// go tool pprof -seconds 60 -http=:8080 http://localhost:6060/debug/pprof/heap
// go tool pprof -http=:8080 http://localhost:6060/debug/pprof/goroutine
// go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=60
