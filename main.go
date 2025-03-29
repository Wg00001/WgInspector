package main

import (
	"WgInspector/adapters/start"
	"WgInspector/entities/config"
	"WgInspector/usecase/client"
	"WgInspector/usecase/db"
	"WgInspector/utils"
	"bufio"
	"context"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"
	"strings"
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

func init() {
	flag.StringVar(&global.ConfigReader, "config-reader", "", "Implementation of config reader")
	flag.StringVar(&global.ConfigParser, "config-parser", "", "Implementation of config parser")
	flag.StringVar(&global.ClientDriver, "client-driver", "", "Type of client driver")
	flag.StringVar(&global.ClientURL, "client-url", "", "Client connection URL")
}

func main() {
	flag.Parse()
	defer cleanup()

	// Initial configuration load
	if err := loadConfiguration(); err != nil {
		log.Fatalf("[ERROR] Initial configuration load failed: %v", err)
	}

	startServices()

	// Start interactive shell
	go commandListener()

	// Signal handling
	go signalHandler()

	// Main loop
	for {
		select {
		case <-restartChan:
			log.Println("[INFO] Restarting services...")
			shutdownServices()
		case <-mainCtx.Done():
			return
		}
	}
}

func loadConfiguration() error {
	// Load existing config
	if err := readConfigFile(); err != nil {
		return fmt.Errorf("config file read error: %w", err)
	}

	// Merge command line inputs
	if hasCommandLineInput() {
		if err := saveConfigFile(); err != nil {
			return fmt.Errorf("config save error: %w", err)
		}
		log.Printf("[INFO] Configuration saved to %s", configPath)
	}
	return nil
}

func startServices() {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()

	// 创建新上下文
	ctx, cancel := context.WithCancel(context.Background())
	mainCtx = ctx
	mainCancel = cancel

	// 初始化系统
	start.Init(func(opt utils.Option) {
		opt["config_reader"] = global.ConfigReader
		opt["config_parser"] = global.ConfigParser
		opt["client_driver"] = global.ClientDriver
		opt["client_url"] = global.ClientURL
		opt.WithOption(global.Option)
	})

	if err := client.Use(global); err != nil {
		log.Printf("[ERROR] Client init: %v", err)
		return
	}

	// 启动客户端（使用新上下文）
	//serviceWg.Add(1)
	go func() {
		//defer serviceWg.Done()
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

	// 触发关闭
	if mainCancel != nil {
		mainCancel()
	}

	// 等待服务停止（最多3秒）
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

	// 强制清理资源
	client.Close()
	db.CloseAll()

	// 重置等待组
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

func commandListener() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nInteractive commands available:")
	fmt.Println("  reload    - Modify configuration")
	fmt.Println("  exit      - Shutdown system")
	fmt.Println("  restart   - Restart services")

	for {
		fmt.Print("\nCommand > ")
		if !scanner.Scan() {
			break
		}

		cmd := strings.TrimSpace(scanner.Text())
		switch cmd {
		case "reload":
			handleReloadCommand()
		case "restart":
			restartChan <- struct{}{}
		case "exit":
			mainCancel()
			return
		default:
			fmt.Println("Invalid command. Available: reload, restart, exit")
		}
	}
}

func handleReloadCommand() {
	newCfg := global

	fmt.Println("\nCurrent configuration:")
	showConfiguration()

	fmt.Println("\nEnter new values (press Enter to keep current):")
	fmt.Printf("Config Reader [%s]: ", newCfg.ConfigReader)
	newCfg.ConfigReader = readInput(newCfg.ConfigReader)

	fmt.Printf("Config Parser [%s]: ", newCfg.ConfigParser)
	newCfg.ConfigParser = readInput(newCfg.ConfigParser)

	fmt.Printf("Client Driver [%s]: ", newCfg.ClientDriver)
	newCfg.ClientDriver = readInput(newCfg.ClientDriver)

	fmt.Printf("Client URL [%s]: ", newCfg.ClientURL)
	newCfg.ClientURL = readInput(newCfg.ClientURL)

	fmt.Print("Save changes? (y/n): ")
	if strings.ToLower(readInput("")) == "y" {
		global = newCfg
		if err := saveConfigFile(); err != nil {
			log.Printf("[ERROR] Save failed: %v", err)
			return
		}
		fmt.Print("Restart to apply changes? (y/n): ")
		if strings.ToLower(readInput("")) == "y" {
			restartChan <- struct{}{}
		}
	}
}

func readInput(defaultVal string) string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal
	}
	return input
}

func showConfiguration() {
	fmt.Printf("Config Reader: %s\n", global.ConfigReader)
	fmt.Printf("Config Parser: %s\n", global.ConfigParser)
	fmt.Printf("Client Driver: %s\n", global.ClientDriver)
	fmt.Printf("Client URL:    %s\n", global.ClientURL)
}

func initializeSystem() {

}

func startClient() error {
	if err := client.Use(global); err != nil {
		return fmt.Errorf("client initialization error: %w", err)
	}
	go client.Listen(mainCtx)
	return nil
}

type InitConfigYaml struct {
	ConfigReader string       `yaml:"config_reader"`
	ConfigParser string       `yaml:"config_parser"`
	ClientDriver string       `yaml:"client_driver"`
	ClientURL    string       `yaml:"client_url"`
	Option       utils.Option `yaml:"option"`
}

func useYaml(initConfig config.InitConfig) InitConfigYaml {
	return InitConfigYaml{
		ConfigReader: initConfig.ConfigReader,
		ConfigParser: initConfig.ConfigParser,
		ClientDriver: initConfig.ClientDriver,
		ClientURL:    initConfig.ClientURL,
		Option:       initConfig.Option,
	}
}
func readConfigFile() error {
	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return saveConfigFile()
		}
		return err
	}
	var unmarshalTarget InitConfigYaml
	err = yaml.Unmarshal(file, &unmarshalTarget)
	if err != nil {
		return err
	}
	global = config.InitConfig{
		ConfigReader: unmarshalTarget.ConfigReader,
		ConfigParser: unmarshalTarget.ConfigParser,
		ClientDriver: unmarshalTarget.ClientDriver,
		ClientURL:    unmarshalTarget.ClientURL,
		Option:       unmarshalTarget.Option,
	}
	return nil
}

func saveConfigFile() error {
	data, err := yaml.Marshal(useYaml(global))
	if err != nil {
		return fmt.Errorf("serialization error: %w", err)
	}

	if err := os.MkdirAll("./app", 0755); err != nil {
		return fmt.Errorf("directory creation error: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("file write error: %w", err)
	}
	return nil
}

func hasCommandLineInput() bool {
	hasInput := false
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "config-reader", "config-parser", "client-driver", "client-url":
			hasInput = true
		}
	})
	return hasInput
}
