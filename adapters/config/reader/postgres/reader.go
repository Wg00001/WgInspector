package postgres

import (
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"WgInspector/utils"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/18
 */

type ConfigReaderPostgre struct {
	*gorm.DB
}

var _ config.Reader = (*ConfigReaderPostgre)(nil)

func (ConfigReaderPostgre) NewReader(option utils.Option) (config.Reader, error) {
	dsn, ok := option["DSN"]
	if !ok {
		return nil, fmt.Errorf("please input DSN\n")
	}
	res, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}
	return ConfigReaderPostgre{DB: res}, nil
}

func (c ConfigReaderPostgre) ReadConfig() error {
	config2.RLock()
	defer config2.RUnlock()
	meta := config.ConfigMeta{}
	c.Table(config.TypeInspector).Select("*").Find(meta.DBs)
	c.Table(config.TypeAgent).Select("*").Find(meta.Agent)
	c.Table(config.TypeDB).Select("*").Find(meta.DBs)
	c.Table(config.TypeTask).Select("*").Find(meta.Tasks)
	c.Table(config.TypeKBase).Select("*").Find(meta.KnowledgeBases)
	c.Table(config.TypeAgentTask).Select("*").Find(meta.AgentTasks)
	c.Table(config.TypeLog).Select("*").Find(meta.Logs)
	c.Table(config.TypeAlert).Select("*").Find(meta.Alerts)
	return config2.SetConfigMeta(meta)
}

func (c ConfigReaderPostgre) SaveConfig(data config.Id) error {
	dataType, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}
	return c.Table(dataType).
		Save(data).
		Where("identity = ?", data.GetIdentity()).
		Error
}

func (c ConfigReaderPostgre) Watch() {
}
