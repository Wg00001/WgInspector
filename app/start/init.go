package start

import (
	"WgInspector/entities/config"
	"WgInspector/usecase/agent"
	"WgInspector/usecase/agent/analyzer"
	"WgInspector/usecase/agent/kbase"
	"WgInspector/usecase/alerter"
	config2 "WgInspector/usecase/config"
	"WgInspector/usecase/db"
	"WgInspector/usecase/logger"
	"WgInspector/usecase/task"
	"WgInspector/usecase/task/cron"
	"WgInspector/utils"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)
import (
	_ "WgInspector/adapters/agent/analyzer/default"
	_ "WgInspector/adapters/agent/analyzer/ollama"
	_ "WgInspector/adapters/agent/analyzer/openai"
	_ "WgInspector/adapters/agent/kbase/chroma"

	_ "WgInspector/adapters/alerter/default"
	_ "WgInspector/adapters/alerter/feishu"

	_ "WgInspector/adapters/client/client_db/pgsql"
	_ "WgInspector/adapters/client/client_db/sqlite"
	_ "WgInspector/adapters/client/websocket"

	_ "WgInspector/adapters/config"

	_ "WgInspector/adapters/logger/default"
	_ "WgInspector/adapters/logger/postgres"

	_ "WgInspector/adapters/cron"

	_ "gorm.io/driver/postgres"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/17
 */

func Init(initConfig config.InitConfig) {
	log.SetFlags(log.LstdFlags)
	gormDB, err := gorm.Open(postgres.Open(initConfig.BaseDSN))
	if err != nil {
		panic(err)
	}
	err = config2.InitReader(gormDB)
	if err != nil {
		panic(fmt.Sprintf("config use fail: %s", err))
	}
	err = config2.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config load fail: %s", err))
	}
	err = InitDB()
	if err != nil {
		log.Printf("db init fail: %s\n", err)
	}

	defer func() {
		if r := recover(); r != nil {
			db.CloseAll()
			panic(r)
		}
	}()
	printErr := func(err error) {
		if err != nil {
			fmt.Printf("!!!!! System init fail !!!!!\n!!!!! Err :%s\n\n", err)
		}
	}

	printErr(InitLogger())
	printErr(InitTask())
	printErr(InitAlert())
	printErr(InitAiConfig())
	printErr(InitAiTask())
	printErr(InitKBase())
	//printErr(client.UseDriver(config.InitConfig{
	//	ClientDriver: "websocket",
	//	ClientURL:    "ws://127.0.0.1:9999",
	//}))
	log.Println("====== System Init Completely ======")
}

func InitOld(optionFuncArr ...utils.OptionFunc) {
	log.SetFlags(log.LstdFlags)

	opt := make(utils.Option)
	opt.With(optionFuncArr...)
	localFileOptFunc(opt)
	//err := config2.UseDriver(config.InitConfig{})
	//if err != nil {
	//	panic(fmt.Sprintf("config use fail: %s", err))
	//}
	err := config2.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config load fail: %s", err))
	}
	err = InitDB()
	if err != nil {
		panic(fmt.Sprintf("db init fail: %s", err))
	}

	defer func() {
		if r := recover(); r != nil {
			db.CloseAll()
			panic(r)
		}
	}()
	printErr := func(err error) {
		if err != nil {
			panic(fmt.Sprintf("!!!!! System init fail !!!!!\n!!!!! Err :%s\n\n", err))
		}
	}

	printErr(InitLogger())
	printErr(InitTask())
	printErr(InitAlert())
	printErr(InitAiConfig())
	printErr(InitAiTask())
	printErr(InitKBase())
	//printErr(client.UseDriver(config.InitConfig{
	//	ClientDriver: "websocket",
	//	ClientURL:    "ws://127.0.0.1:9999",
	//}))
	log.Println("====== System Init Completely ======")
}

func InitDB() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.DBs {
		err := db.Use(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitLogger() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.Logs {
		err := logger.Use(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitTask() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.Tasks {
		t := task.NewInspTask(v)
		err := cron.AddTask(&t)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitAlert() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.Alerts {
		err := alerter.Use(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitAiConfig() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.Agents {
		err := analyzer.Register(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitAiTask() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.AgentTasks {
		err := cron.AddTask(agent.NewTask(v))
		if err != nil {
			return err
		}
	}
	return nil
}

func InitKBase() error {
	config2.RLock()
	defer config2.RUnlock()
	for _, v := range config2.Meta.KBases {
		err := kbase.Use(v)
		if err != nil {
			return err
		}
	}
	return nil
}
