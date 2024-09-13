package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sync"
)

// DBNameArray 定义 DBNameArray 类型，允许 dbname 字段为字符串或字符串数组
type DBNameArray []string

// UnmarshalYAML 实现自定义解析逻辑
func (d *DBNameArray) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var single string
	var array []string
	// 尝试将 YAML 数据解析为单个字符串
	if err := unmarshal(&single); err == nil {
		*d = []string{single}
		return nil
	}
	// 尝试将 YAML 数据解析为字符串数组
	if err := unmarshal(&array); err == nil {
		*d = array
		return nil
	}
	return errors.New("failed to unmarshal dbname field")
}

// MySQLConfig 结构体定义
type MySQLConfig struct {
	Host     string      `yaml:"host"`
	Port     int         `yaml:"port"`
	User     string      `yaml:"user"`
	Password string      `yaml:"password"`
	DBName   DBNameArray `yaml:"dbname"`
	Charset  string      `yaml:"charset"`
	Timeout  int         `yaml:"timeout"`
}

// Config 结构体定义
type Config struct {
	ResultDB  *MySQLConfig  `yaml:"result_db"`
	TargetDB  []MySQLConfig `yaml:"target_db"`
	LogDir    string        `yaml:"log_dir"`
	FeishuURL string        `yaml:"feishu_url"`
}

// 单例模式相关变量
var (
	config *Config
	once   sync.Once
)

// LoadConfig 初始化配置
func LoadConfig(configPath string) error {
	// 打开 YAML 文件
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	// 创建一个 Config 对象
	// 使用 yaml 解码文件内容到 Config 对象中
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	return nil
}

func InitConfig() error {
	if config == nil {
		return errors.New("config not loaded")
	}
	if config.LogDir != "" {
		file, err := os.OpenFile(config.LogDir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		// 设置日志输出到文件
		log.SetOutput(file)
	}
	if config.FeishuURL != "" {

	}
	if config.ResultDB != nil {

	}
	if config.TargetDB != nil {

	}
	return nil
}

// GetConfigInstance 单例模式下的配置文件解析函数
func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}
	// 使用 sync.Once 确保解析操作只执行一次
	return nil, errors.New("config not loaded")
}
