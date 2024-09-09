package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

// Config 结构体，表示配置文件中变量，与config.yaml文件里的配置对应
type Config struct {
	ResDB     MySQLConfig   `yaml:"result_db"`
	TargetDB  []MySQLConfig `yaml:"target_db"`
	LogDir    string        `yaml:"log_dir"`
	FeishuURL string        `yaml:"feishu_url"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`     // 数据库主机名
	Port     int    `yaml:"port"`     // 数据库端口号
	User     string `yaml:"user"`     // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
	DBName   string `yaml:"dbname"`   // 数据库名称
	Charset  string `yaml:"charset"`  // 数据库字符集
	Timeout  int    `yaml:"timeout"`  // 连接超时（秒）
}

// LoadConfig 初始化配置
func LoadConfig(configPath string) (Config, error) {
	// 读取配置文件内容
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	configMap := make(map[string]interface{})
	err = yaml.Unmarshal(configFile, &configMap)
	// 如果解析配置文件过程中发生错误，返回nil和错误信息
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	// 返回解析后的Config结构体对象和nil
	return unmarshalConfig(configMap)
}

func unmarshalConfig(configMap map[string]interface{}) (Config, error) {
	res := Config{
		ResDB:    MySQLConfig{},
		TargetDB: []MySQLConfig{},
		//todo:捕获panic
		LogDir:    configMap["log_dir"].(string),
		FeishuURL: configMap["feishu_url"].(string),
	}
	return res, nil
}
