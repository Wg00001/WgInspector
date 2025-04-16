package test

import (
	_ "WgInspector/adapters/agent/kbase/chroma"
	start2 "WgInspector/app/start"
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	agent2 "WgInspector/usecase/agent"
	"WgInspector/usecase/agent/kbase"
	"context"
	"fmt"
	"testing"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/6
 */

func TestFullAgentRAU(t *testing.T) {
	start2.SetLocalConfigReaderOption("../../app/config", "local_file")
	start2.InitOld()
	start2.Run(context.Background())
}

func TestAgentRAU(t *testing.T) {
	start2.SetLocalConfigReaderOption("../../app/config", "yaml")
	start2.InitOld()
	//start.Run(context.Background())
	nt := agent2.NewTask(&config.AgentTaskConfig{
		Identity: "agent_test",
		Cron: &config.Cron{
			AtTime: []string{time.Now().Add(time.Second * 3).Format(time.TimeOnly)},
		},
		LogID: "1",
		LogFilter: config.LogFilter{
			StartTime: func() time.Time {
				parse, _ := time.Parse(time.DateOnly, "2025-03-05")
				return parse
			}(),
			EndTime:   time.Now(),
			TaskNames: nil,
			DBNames:   nil,
			TaskIDs:   nil,
			InspNames: nil,
		},
		AlertID:      "3",
		KBase:        []config.Identity{"chroma"},
		KBaseResults: 3,
		KBaseMaxLen:  100000,
	})
	//cron.AddTask(nt)
	//start.Run(context.Background())
	err := nt.Do(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func TestKBase(t *testing.T) {
	start2.SetLocalConfigReaderOption("../../app/config", "yaml")
	start2.InitOld()

	// 初始化测试配置
	cfg := config.KnowledgeBaseConfig{
		Identity: "unit-test-kb",
		Driver:   "chroma",
		Value: map[string]interface{}{
			"path":       "http://localhost:8000",
			"collection": "ollama_embedding",
			"embedding":  "ollama",
			"baseurl":    "http://127.0.0.1:11434",
			"model":      "deepseek-r1:7b",
			//"tenant":     "ollama_embedding",
		},
	}

	// 测试初始化
	err := kbase.Use(cfg)
	if err != nil {
		t.Fatal(err)
	}
	kb := kbase.Get("unit-test-kb")
	fmt.Printf("%+v\n", kb)

	// 准备测试数据
	testDocs := []agent.Document{
		{
			ID:      "doc5",
			Content: "Go语言并发编程指南2",
			//Embedding: []float32{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{"author": "张三", "timestamp": time.Now().Unix()},
		},
		{
			ID:      "doc6",
			Content: "分布式系统设计原则2",
			//Embedding: []float32{0.4, 0.5, 0.6},
			Metadata: map[string]interface{}{"year": 2023, "timestamp": time.Now().Unix()},
		},
	}

	t.Run("WriteDocuments", func(t *testing.T) {
		// 正常写入测试
		if err := kb.WriteIn(testDocs); err != nil {
			t.Fatalf("WriteIn failed: %v", err)
		}

		// 写入空数据测试
		t.Run("EmptyDocuments", func(t *testing.T) {
			err := kb.WriteIn(nil)
			if err == nil {
				t.Error("Expected error for empty documents")
			}
		})

		// 无效文档测试
		t.Run("InvalidDocument", func(t *testing.T) {
			invalidDocs := []agent.Document{{ID: ""}}
			err := kb.WriteIn(invalidDocs)
			if err == nil {
				t.Error("Expected error for invalid document")
			}
		})
	})

	t.Run("Search", func(t *testing.T) {
		// 正常搜索测试
		t.Run("BasicSearch", func(t *testing.T) {
			results, err := kb.Search(
				agent.NewQueryData().
					WithKeyWords("并发").
					WithMetaData(map[string]string{"author": "张三"}))
			//WithMetaData(map[string]string{}))
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			fmt.Println(results)
			for _, v := range results {
				fmt.Println(v)
			}
		})

		// 多查询测试
		t.Run("MultipleQueries", func(t *testing.T) {
			results, err := kb.Search(
				agent.NewQueryData().
					WithKeyWords("并发").
					WithMetaData(map[string]string{"author": "张三"}))
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println(results)
		})

		// 边界测试
		t.Run("BoundaryConditions", func(t *testing.T) {
			// 零结果测试
			results, err := kb.Search(agent.QueryData{KeyWords: []string{"无关内容"}})
			if err == nil {
				t.Error("Expected error for empty query")
			}
			fmt.Println(results)
			// 无效查询测试
			results, err = kb.Search(agent.QueryData{KeyWords: []string{""}})

			if err == nil {
				t.Error("Expected error for empty query")
			}
			fmt.Println(results)
		})
	})

	t.Run("Embedding", func(t *testing.T) {
		// 正常生成测试
		emb, err := kb.Embedding("生成向量")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(emb)
	})

	// 清理测试数据
	t.Cleanup(func() {
		if memKb, ok := kb.(interface{ Reset() }); ok {
			memKb.Reset()
		}
	})
}
