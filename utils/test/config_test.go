package test

import (
	"WgInspector/adapters/config"
	"WgInspector/app/start"
	config2 "WgInspector/entities/config"
	config3 "WgInspector/usecase/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/23
 */

func TestCreate(t *testing.T) {
	gormDB, err := gorm.Open(postgres.Open("host=127.0.0.1 port=5432 user=postgres dbname=postgres password=03719 sslmode=disable"))
	if err != nil {
		panic(err)
	}
	db, _ := config.ConfigReaderPostgre{}.NewReader(gormDB)

	data := config2.DBConfig{
		Identity: config2.Identity{
			Name: "9",
		},
		Driver: "9",
		DSN:    "9",
	}

	id, err := db.SaveConfig(data)
	fmt.Println(id)
	fmt.Println(data)
	if err != nil {
		panic(err)
	}
	fmt.Println()
}

func TestGet(t *testing.T) {
	start.Init(config2.InitConfig{
		BaseDSN:   "host=127.0.0.1 port=5432 user=postgres dbname=postgres password=03719 sslmode=disable",
		ClientURL: "ws://0.0.0.0:9999/",
	})
	res, err := config3.Get(config3.Key{
		ConfigType: config2.TypeDB,
		Identity: config2.Identity{
			ID:   14,
			Name: "2",
		},
	})
	fmt.Println(res, err)

	res2, err := config3.GetWithType[config2.DBConfig](config3.Key{
		ConfigType: config2.TypeDB,
		Identity: config2.Identity{
			ID:   14,
			Name: "2",
		},
	})
	fmt.Println(res2, err)
}
