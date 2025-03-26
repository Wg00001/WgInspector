package config

import (
	"WgInspector/entities/config"
	"fmt"
	"log"
	"sync"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/4
 */

func Open(driverName string, option map[string]string) error {
	//todo: option封装
	reader, err := GetReader(driverName)
	if err != nil {
		return err
	}
	reader, err = reader.NewReader(option)
	if err != nil {
		return err
	}
	err = reader.ReadConfig()
	if err != nil {
		return err
	}
	reader.SaveIntoConfig()
	log.Println("config initiated...")
	return nil
}

var (
	readerDrivers  = make(map[string]config.Reader)
	readerDriverMu sync.RWMutex
)

func RegisterReader(name string, driver config.Reader) {
	readerDriverMu.Lock()
	defer readerDriverMu.Unlock()
	if readerDrivers == nil {
		panic("config: readerDrivers map is nil")
	}
	if _, dup := readerDrivers[name]; dup {
		panic("config: Register called twice for driver " + name)
	}
	readerDrivers[name] = driver
}

func GetReader(name string) (config.Reader, error) {
	readerDriverMu.RLock()
	defer readerDriverMu.RUnlock()
	res, ok := readerDrivers[name]
	if !ok {
		return nil, fmt.Errorf("config: get driver fail %s\n", name)
	}
	return res, nil
}

var (
	parserDrivers  = make(map[string]config.Parser)
	parserDriverMu sync.RWMutex
)

func RegisterParser(name string, driver config.Parser) {
	parserDriverMu.Lock()
	defer parserDriverMu.Unlock()
	if parserDrivers == nil {
		panic("config: readerDrivers map is nil")
	}
	if _, dup := parserDrivers[name]; dup {
		panic("config: Register called twice for driver " + name)
	}
	parserDrivers[name] = driver
}

func GetParser(name string) (config.Parser, error) {
	parserDriverMu.RLock()
	defer parserDriverMu.RUnlock()
	res, ok := parserDrivers[name]
	if !ok {
		return nil, fmt.Errorf("config: get driver fail %s\n", name)
	}
	return res, nil
}
