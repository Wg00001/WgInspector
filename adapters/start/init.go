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
	"github.com/wg00001/wgo-sdk/wg"
	"log"
)
import (
	_ "WgInspector/adapters/agent/analyzer/default"
	_ "WgInspector/adapters/agent/analyzer/ollama"
	_ "WgInspector/adapters/agent/analyzer/openai"
	_ "WgInspector/adapters/agent/kbase/chroma"

	_ "WgInspector/adapters/alerter/default"
	_ "WgInspector/adapters/alerter/empty"
	_ "WgInspector/adapters/alerter/feishu"

	_ "WgInspector/adapters/client"
	_ "WgInspector/adapters/client/websocket"

	_ "WgInspector/adapters/config/parser/json"
	_ "WgInspector/adapters/config/parser/yaml"
	_ "WgInspector/adapters/config/reader/etcd"
	_ "WgInspector/adapters/config/reader/local_file"

	_ "WgInspector/adapters/logger/default"
	_ "WgInspector/adapters/logger/postgres"

	_ "WgInspector/adapters/cron"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/17
 */

func Init(initConfig config.InitConfig) {
	log.SetFlags(log.LstdFlags)
	err := config2.UseDriver(initConfig)
	if err != nil {
		panic(fmt.Sprintf("config use fail: %s", err))
	}
	err = config2.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("config load fail: %s", err))
	}
	err = InitDB()
	if err != nil {
		fmt.Printf("db init fail: %s\n", err)
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
	//printErr(client.Use(config.InitConfig{
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
	err := config2.UseDriver(config.InitConfig{})
	if err != nil {
		panic(fmt.Sprintf("config use fail: %s", err))
	}
	err = config2.LoadConfig()
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
	//printErr(client.Use(config.InitConfig{
	//	ClientDriver: "websocket",
	//	ClientURL:    "ws://127.0.0.1:9999",
	//}))
	log.Println("====== System Init Completely ======")
}

func InitDB() error {
	config2.RLock()
	defer config2.RUnlock()
	dbConfigs := wg.MapToValueSlice(config2.Index.DB)
	for _, v := range dbConfigs {
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
	logConfigs := wg.MapToValueSlice(config2.Index.Log)
	for _, v := range logConfigs {
		err := logger.Use(*v)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitTask() error {
	config2.RLock()
	defer config2.RUnlock()
	taskConfigs := wg.MapToValueSlice(config2.Index.Task)
	for _, v := range taskConfigs {
		t, err := task.NewTask(v)
		if err != nil {
			return err
		}
		err = task.Register(t)
		if err != nil {
			return err
		}
		cron.AddTask(t)
	}
	return nil
}

func InitAlert() error {
	config2.RLock()
	defer config2.RUnlock()
	alertConfigs := wg.MapToValueSlice(config2.Index.Alert)
	for _, v := range alertConfigs {
		err := alerter.Use(*v)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitAiConfig() error {
	config2.RLock()
	defer config2.RUnlock()
	return analyzer.Use(*config2.Index.Agent)
}

func InitAiTask() error {
	config2.RLock()
	defer config2.RUnlock()
	aiTasks := wg.MapToValueSlice(config2.Index.AgentTask)
	for _, v := range aiTasks {
		cron.AddTask(agent.NewTask(v))
	}
	return nil
}

func InitKBase() error {
	config2.RLock()
	defer config2.RUnlock()
	kbaseConfig := wg.MapToValueSlice(config2.Index.KBase)
	for _, v := range kbaseConfig {
		err := kbase.Use(*v)
		if err != nil {
			return err
		}
	}
	return nil
}
