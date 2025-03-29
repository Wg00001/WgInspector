package cli

import (
	"WgInspector/entities/config"
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

type CLI struct {
	restartChan chan struct{}
	cancelFunc  func()
	config      *config.InitConfig
}

func NewCLI(restartChan chan struct{}, cancelFunc func(), config *config.InitConfig) *CLI {
	return &CLI{
		restartChan: restartChan,
		cancelFunc:  cancelFunc,
		config:      config,
	}
}

func (c *CLI) StartCommandListener() {
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
			c.handleReloadCommand()
		case "restart":
			c.restartChan <- struct{}{}
		case "exit":
			c.cancelFunc()
			return
		default:
			fmt.Println("Invalid command. Available: reload, restart, exit")
		}
	}
}

func (c *CLI) handleReloadCommand() {
	newCfg := *c.config

	fmt.Println("\nCurrent configuration:")
	c.showConfiguration()

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
		*c.config = newCfg
		if err := saveConfigFile(*c.config); err != nil {
			log.Printf("[ERROR] Save failed: %v", err)
			return
		}
		fmt.Print("Restart to apply changes? (y/n): ")
		if strings.ToLower(readInput("")) == "y" {
			c.restartChan <- struct{}{}
		}
	}
}

func (c *CLI) showConfiguration() {
	fmt.Printf("Config Reader: %s\n", c.config.ConfigReader)
	fmt.Printf("Config Parser: %s\n", c.config.ConfigParser)
	fmt.Printf("Client Driver: %s\n", c.config.ClientDriver)
	fmt.Printf("Client URL:    %s\n", c.config.ClientURL)
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

type InitConfigYaml struct {
	ConfigReader string            `yaml:"config_reader"`
	ConfigParser string            `yaml:"config_parser"`
	ClientDriver string            `yaml:"client_driver"`
	ClientURL    string            `yaml:"client_url"`
	Option       map[string]string `yaml:"option"`
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

func saveConfigFile(cfg config.InitConfig) error {
	data, err := yaml.Marshal(useYaml(cfg))
	if err != nil {
		return fmt.Errorf("serialization error: %w", err)
	}

	if err := os.MkdirAll("./app", 0755); err != nil {
		return fmt.Errorf("directory creation error: %w", err)
	}

	if err := os.WriteFile("./app/init_config.yaml", data, 0644); err != nil {
		return fmt.Errorf("file write error: %w", err)
	}
	return nil
}
