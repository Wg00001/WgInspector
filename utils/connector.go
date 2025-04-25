package utils

import (
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/25
 */

func ConnectGormDB(driver, DSN string, config ...gorm.Option) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch driver {
	case "mysql":
		dialector = mysql.Open(DSN)
	case "postgres":
		dialector = postgres.Open(DSN)
	default:
		return nil, errors.New("connect gorm db fail: unsupported driver: " + driver)
	}
	db, err := gorm.Open(dialector, config...)
	if err != nil {
		return nil, err
	}
	return db, nil
}
