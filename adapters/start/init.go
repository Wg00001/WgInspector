package start

import (
	"PgInspector/adapters/cron"
	"PgInspector/usecase/agent"
	"PgInspector/usecase/agent/analyzer"
	"PgInspector/usecase/agent/kbase"
	"PgInspector/usecase/alerter"
	config2 "PgInspector/usecase/config"
	"PgInspector/usecase/db"
	"PgInspector/usecase/logger"
	"PgInspector/usecase/task"
	"fmt"
	"github.com/wg00001/wgo-sdk/wg"
	"log"
)
import (
	_ "PgInspector/adapters/agent/analyzer/default"
	_ "PgInspector/adapters/agent/analyzer/ollama"
	_ "PgInspector/adapters/agent/analyzer/openai"
	_ "PgInspector/adapters/alerter/default"
	_ "PgInspector/adapters/alerter/empty"
	_ "PgInspector/adapters/alerter/feishu"
	_ "PgInspector/adapters/client/websocket"
	_ "PgInspector/adapters/config/parser/yaml"
	_ "PgInspector/adapters/config/reader/local_file"
	_ "PgInspector/adapters/logger/default"
	_ "PgInspector/adapters/logger/postgres"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/17
 */

func Init() {
	log.SetFlags(log.LstdFlags)

	//config.InitConfig(config.NewReader(pConfigType, pFilePath))
	err := config2.Open(pConfigType, map[string]string{
		"filepath": pFilePath,
	})
	if err != nil {
		panic(fmt.Sprintf("config open fail: %s", err))
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

	cron.Init()
	printErr(InitLogger())
	printErr(InitTask())
	printErr(InitAlert())
	printErr(InitAiConfig())
	printErr(InitAiTask())
	printErr(InitKBase())
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

func InitCron() error {
	cron.Init()
	return nil
}
