package db

import (
	"WgInspector/entities/config"
	"WgInspector/entities/db"
	config2 "WgInspector/usecase/config"
	"fmt"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/10
 */

var pool = sync.Map{}

func Register(sqlDB *db.SqlDB) error {
	if _, ok := pool.Load(sqlDB.Config.Identity); ok {
		return fmt.Errorf("sql db is already exist, db name: %s\n", sqlDB.Config.Identity)
	}
	pool.Store(sqlDB.Config.Identity, sqlDB)
	return nil
}

func Get(id config.Identity) (*db.SqlDB, error) {
	dbc, err := config2.Get(config2.Key{
		ConfigType: config.TypeDB,
		Identity:   id,
	})
	if err != nil {
		return nil, err
	}
	dbconfig, ok := dbc.(config.DBConfig)
	if !ok {
		return nil, fmt.Errorf("db config is not config.DBConfig")
	}
	return Build(dbconfig)
}

func GetList(ids []config.Identity) ([]*db.SqlDB, error) {
	res := make([]*db.SqlDB, 0, len(ids))
	for _, id := range ids {
		dbc, err := Get(id)
		if err != nil {
			return nil, err
		}
		res = append(res, dbc)
	}
	return res, nil
}

func ConnList(configs []config.DBConfig) ([]*db.SqlDB, error) {
	res := make([]*db.SqlDB, 0, len(configs))
	for _, config := range configs {
		build, err := Build(config)
		if err != nil {
			return nil, err
		}
		res = append(res, build)
	}
	return res, nil
}

func Close(arg config.Identity) error {
	if val, ok := pool.LoadAndDelete(arg); ok {
		err := val.(*db.SqlDB).Close()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("DB not exist")
	}
	return nil
}

func CloseAll() error {
	var err error
	pool.Range(func(key, value any) bool {
		er := Close(key.(config.Identity))
		if er != nil {
			err = er
		}
		return true
	})
	return err
}

func GetDriverList() []string {
	return []string{"postgres", "mysql"}
}
