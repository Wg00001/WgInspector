package config

import (
	"WgInspector/entities/config"
	"gorm.io/gorm"
	"log"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/4
 */

var (
	reader         config.Reader
	readerDriverMu sync.RWMutex
)

func InitReader(db *gorm.DB) error {
	r, err := reader.NewReader(db)
	if err != nil {
		return err
	}
	readerDriverMu.Lock()
	defer readerDriverMu.Unlock()
	reader = r
	return nil
}

func UseDriver(r config.Reader) {
	readerDriverMu.Lock()
	defer readerDriverMu.Unlock()
	reader = r
}

func LoadConfig() error {
	meta, err := reader.ReadConfig()
	if err != nil {
		return err
	}
	err = SetConfigMeta(meta)
	if err != nil {
		return err
	}
	log.Println("config initiated...")
	return nil
}
