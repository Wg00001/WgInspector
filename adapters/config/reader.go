package config

import (
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"gorm.io/gorm"
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
	return ConfigReaderPostgre{DB: db}, nil
}

func (c ConfigReaderPostgre) ReadConfig() error {
	config2.RLock()
	defer config2.RUnlock()
	meta := config.MetaConfig{}
	c.Table(config.TypeInspector).Select("*").Find(meta.InspNodes)
	c.Table(config.TypeAgent).Select("*").Find(meta.Agents)
	c.Table(config.TypeDB).Select("*").Find(meta.DBs)
	c.Table(config.TypeTask).Select("*").Find(meta.Tasks)
	c.Table(config.TypeKBase).Select("*").Find(meta.KBases)
	c.Table(config.TypeAgentTask).Select("*").Find(meta.AgentTasks)
	c.Table(config.TypeLog).Select("*").Find(meta.Logs)
	c.Table(config.TypeAlert).Select("*").Find(meta.Alerts)
	return config2.SetConfigMeta(meta)
}

// SaveConfig 创建或更新
func (c ConfigReaderPostgre) SaveConfig(data config.Id) error {
	dataType, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}
	return c.Table(dataType).
		Save(data).
		//Where("identity = ?", data.GetIdentity()).
		Error
}

func (c ConfigReaderPostgre) DeleteConfig(data config.Id) error {
	dataType, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}
	return c.Table(dataType).
		Delete(data).
		Error
}

func (c ConfigReaderPostgre) Watch() {
	//废弃：智能体对于配置的更新也走配置中心
}
