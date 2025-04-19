package main

import (
	"WgInspector/app/start"
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"WgInspector/usecase/db"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	global       config.InitConfig
	configPath   = "./app/init_config.yaml"
	mainCtx      context.Context
	mainCancel   context.CancelFunc
	restartChan  = make(chan struct{}, 1)
	serviceWg    sync.WaitGroup
	serviceMutex sync.Mutex
)

func main() {
	flag.Parse()
	defer cleanup()

	startServices()

	// Signal handling
	go signalHandler()

	// Main loop
	for {
		select {
		case <-restartChan:
			log.Println("[INFO] Restarting services...")
			shutdownServices()
			startServices()
		case <-mainCtx.Done():
			return
		}
	}
}

func startServices() {
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

func cleanup() {
	shutdownServices()
	log.Println("[INFO] === System resources released ===")
}

func shutdownServices() {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()

	if mainCancel != nil {
		mainCancel()
	}

	done := make(chan struct{})
	go func() {
		serviceWg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		log.Println("[WARN] Force shutdown due to timeout")
	}

	client.Close()
	db.CloseAll()

	serviceWg = sync.WaitGroup{}
}

func signalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-sigChan
		log.Printf("[INFO] Received signal: %v", sig)
		mainCancel()
		return
	}
}
