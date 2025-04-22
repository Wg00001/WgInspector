package config

import (
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"WgInspector/utils"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"reflect"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/18
 */

func init() {
	config2.UseDriver(ConfigReaderPostgre{})
}

type ConfigReaderPostgre struct {
	*gorm.DB
}

var _ config.Reader = (*ConfigReaderPostgre)(nil)

func (ConfigReaderPostgre) NewReader(db *gorm.DB) (config.Reader, error) {
	err := db.AutoMigrate(
		&config.DBConfig{},
		&config.AgentConfig{},
		&config.AlertConfig{},
		&config.LogConfig{},
		&config.TaskConfig{},
		&config.AgentTaskConfig{},
		&config.KnowledgeBaseConfig{},
		&config.InspNode{},
	)
	if err != nil {
		return nil, err
	}
	db.Logger = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	return ConfigReaderPostgre{DB: db.Debug()}, nil
}

func (c ConfigReaderPostgre) ReadConfig() error {
	meta := config.MetaConfig{}
	c.Table(config.TypeAgent).Select("*").Find(&meta.Agents)
	c.Table(config.TypeDB).Select("*").Find(&meta.DBs)
	c.Table(config.TypeTask).Select("*").Find(&meta.Tasks)
	c.Table(config.TypeKBase).Select("*").Find(&meta.KBases)
	c.Table(config.TypeAgentTask).Select("*").Find(&meta.AgentTasks)
	c.Table(config.TypeLog).Select("*").Find(&meta.Logs)
	c.Table(config.TypeAlert).Select("*").Find(&meta.Alerts)
	c.Table(config.TypeInspector).Select("*").Find(&meta.InspNodes)
	return config2.SetConfigMeta(meta)
}

// SaveConfig 创建或更新
func (c ConfigReaderPostgre) SaveConfig(data config.Id) error {
	dataType, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}

	// 将结构体转换为map
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// 创建map
	mapData := make(map[string]interface{})
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		// 跳过嵌入字段Identity，单独处理
		if field.Anonymous && field.Type.Name() == "Identity" {
			identityValue := value.Field(i)
			// 获取Identity的Name字段
			if nameField := identityValue.FieldByName("Name"); nameField.IsValid() {
				mapData["name"] = nameField.String()
			}
			if nameField := identityValue.FieldByName("ID"); nameField.IsValid() && nameField.Int() != 0 {
				mapData["id"] = nameField.Int()
			}
			continue
		}
		mapData[utils.ToSnakeCase(field.Name)] = value.Field(i).Interface()
	}
	fmt.Printf("Data type: %T\n", data)

	if data.GetIdentity().ID == 0 {
		return c.Table(dataType).
			Create(mapData).
			Error
	} else {
		return c.Table(dataType).
			Where("id = ?", data.GetIdentity().ID).
			Updates(mapData).
			Error
	}
}

func (c ConfigReaderPostgre) DeleteConfig(data config.Id) error {
	dataType, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}
	return c.Table(dataType).
		Where("id = ?", data.GetIdentity().ID).
		Delete(data).
		Error
}

func (c ConfigReaderPostgre) Watch() {
	//废弃：智能体对于配置的更新也走配置中心
}
