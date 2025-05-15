package config

import (
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
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
		&config.InspConfig{},
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

func (c ConfigReaderPostgre) ReadConfig(configTypes ...string) (config.MetaConfig, error) {
	meta := config.MetaConfig{}
	fMap := map[string]func() error{
		config.TypeDB: func() error {
			return c.Table(config.TypeDB).Select("*").Find(&meta.DBs).Error
		},
		config.TypeTask: func() error {
			return c.Table(config.TypeTask).Preload("TargetDB").Preload("Todo").Preload("NotTodo").Select("*").Find(&meta.Tasks).Error
		},
		config.TypeKBase: func() error {
			return c.Table(config.TypeKBase).Select("*").Find(&meta.KBases).Error
		},
		config.TypeAgentTask: func() error {
			return c.Table(config.TypeAgentTask).Preload("KBase").Select("*").Find(&meta.AgentTasks).Error
		},
		config.TypeLog: func() error {
			return c.Table(config.TypeLog).Select("*").Find(&meta.Logs).Error
		},
		config.TypeAlert: func() error {
			return c.Table(config.TypeAlert).Select("*").Find(&meta.Alerts).Error
		},
		config.TypeAgent: func() error {
			return c.Table(config.TypeAgent).Select("*").Find(&meta.Agents).Error
		},
		config.TypeInspector: func() error {
			return c.Table(config.TypeInspector).Select("*").Find(&meta.InspNodes).Error
		},
	}

	var executeFuncs []func() error
	if len(configTypes) == 0 {
		for _, fn := range fMap {
			executeFuncs = append(executeFuncs, fn)
		}
	} else {
		for _, ct := range configTypes {
			fn, ok := fMap[ct]
			if !ok {
				return meta, fmt.Errorf("config reader: unknown config type: %s", ct)
			}
			executeFuncs = append(executeFuncs, fn)
		}
	}

	for _, fn := range executeFuncs {
		if err := fn(); err != nil {
			return meta, err
		}
	}
	return meta, nil
}

// SaveConfig 创建或更新
func (c ConfigReaderPostgre) SaveConfig(data config.Id) (int64, error) {
	switch v := data.(type) {
	case config.DBConfig:
		res, err := save[config.DBConfig](c.DB, v)
		return res.ID, err
	case config.LogConfig:
		res, err := save[config.LogConfig](c.DB, v)
		return res.ID, err
	case config.AlertConfig:
		res, err := save[config.AlertConfig](c.DB, v)
		return res.ID, err
	case config.TaskConfig:
		res, err := save[config.TaskConfig](c.DB, v)
		return res.ID, err
	case config.AgentConfig:
		res, err := save[config.AgentConfig](c.DB, v)
		return res.ID, err
	case config.AgentTaskConfig:
		res, err := save[config.AgentTaskConfig](c.DB, v)
		return res.ID, err
	case config.KnowledgeBaseConfig:
		res, err := save[config.KnowledgeBaseConfig](c.DB, v)
		return res.ID, err
	case config.InspConfig:
		res, err := save[config.InspConfig](c.DB, v)
		return res.ID, err
	default:
		return 0, fmt.Errorf("config_reader - save: unknown config type: %T", data)
	}
}

func save[T config.ConfigType](db *gorm.DB, data config.Id) (T, error) {
	res := config.Turn[T](data)
	var err error
	if data.GetIdentity().ID == 0 {
		err = db.Create(&res).Error
	} else {
		err = db.Where("id = ?", data.GetIdentity().ID).
			Updates(&res).
			Error
	}
	return res, err
}

func (c ConfigReaderPostgre) DeleteConfig(data config.Id) error {
	dataType, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}

	// For types with many2many relationships, clear associations first
	switch data.(type) {
	case config.TaskConfig, *config.TaskConfig:
		var task config.TaskConfig
		if err := c.DB.Preload("TargetDB").Preload("Todo").Preload("NotTodo").First(&task, data.GetIdentity().ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound { // If record not found, it might have been already deleted or associations cleared
				return nil
			}
			return fmt.Errorf("failed to load task config for deletion: %w", err)
		}
		if err := c.DB.Model(&task).Association("TargetDB").Clear(); err != nil {
			return fmt.Errorf("failed to clear TargetDB association for TaskConfig: %w", err)
		}
		if err := c.DB.Model(&task).Association("Todo").Clear(); err != nil {
			return fmt.Errorf("failed to clear Todo association for TaskConfig: %w", err)
		}
		if err := c.DB.Model(&task).Association("NotTodo").Clear(); err != nil {
			return fmt.Errorf("failed to clear NotTodo association for TaskConfig: %w", err)
		}
	case config.AgentTaskConfig, *config.AgentTaskConfig:
		var agentTask config.AgentTaskConfig
		if err := c.DB.Preload("KBase").First(&agentTask, data.GetIdentity().ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return fmt.Errorf("failed to load agent task config for deletion: %w", err)
		}
		if err := c.DB.Model(&agentTask).Association("KBase").Clear(); err != nil {
			return fmt.Errorf("failed to clear KBase association for AgentTaskConfig: %w", err)
		}
	}

	return c.Table(dataType).
		Where("id = ?", data.GetIdentity().ID).
		Delete(data).
		Error
}

func (c ConfigReaderPostgre) Watch() {
	//废弃：智能体对于配置的更新也走配置中心
}
