package db

import (
	"WgInspector/entities/config"
	"WgInspector/entities/db"
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

func Get(id config.Identity) *db.SqlDB {
	if val, ok := pool.Load(id); ok {
		return val.(*db.SqlDB)
	}
	return &db.SqlDB{Err: fmt.Errorf("db config is nil")}
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
