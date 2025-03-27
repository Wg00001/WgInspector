package test

import (
	"WgInspector/adapters/agent/analyzer/ollama"
	"WgInspector/adapters/cron"
	"WgInspector/adapters/start"
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	ai2 "WgInspector/usecase/agent"
	"WgInspector/usecase/agent/analyzer"
	config2 "WgInspector/usecase/config"
	"context"
	"fmt"
	"testing"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/2/26
 */

func TestOllamaApi(t *testing.T) {
	a := ollama.AnalyzerOllama{
		Driver:      "ollama",
		Url:         "http://127.0.0.1:11434",
		ApiKey:      "",
		Model:       "deepseek-r1:7b",
		Temperature: 0.5,
	}
	analyze, err := a.Analyze(&agent.AnalyzeContent{
		UserMsg: "你好，你是什么模型",
	})
	if err != nil {
		return
	}
	fmt.Println(analyze)
}

func TestAiTask(t *testing.T) {
	//start.SetLocalConfigReaderOption("../../app/config", "yaml")
	fmt.Println(config2.Open("yaml", map[string]string{
		"filepath": "../../app/config",
	}))
	fmt.Println(start.InitDB())
	cron.Init()
	fmt.Println(start.InitLogger())
	fmt.Println(start.InitAlert())
	//fmt.Println(start.InitAi())
	err := analyzer.Use(*config2.Index.Agent)
	if err != nil {
		fmt.Println(err)
	}
	tsk := ai2.NewTask(&config.AgentTaskConfig{
		Identity: "1",
		Cron: &config.Cron{
			Duration: time.Second * 10,
		},
		LogID: "1",
		LogFilter: config.LogFilter{
			StartTime: time.Now().AddDate(0, 0, -3),
			InspNames: []config.Identity{"1"},
		},
		AlertID: "3",
	})
	fmt.Println(tsk.Do(context.Background()))
}

func TestAi(t *testing.T) {
	start.SetLocalConfigReaderOption("../../app/config", "yaml")
	start.Init()
	start.Run(context.TODO())
}
