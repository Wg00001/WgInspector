package db

import (
	"WgInspector/entities/config"
	"WgInspector/entities/db"
	"fmt"
	"log"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/15
 */

func Use(dbConfig *config.DBConfig) error {
	sqlDB, err := Build(dbConfig)
	if err != nil {
		return err
	}
	return Register(sqlDB)
}

func Build(dbConfig *config.DBConfig) (*db.SqlDB, error) {
	if dbConfig == nil {
		return nil, fmt.Errorf("db config is nil")
	}
	cur := &db.SqlDB{Config: dbConfig}
	err := cur.Connect()
	if err != nil {
		return nil, err
	}
	if err := cur.Ping(); err != nil {
		return nil, err
	} else {
		log.Println("	db: connected - " + dbConfig.Identity)
	}
	return cur, nil
}
