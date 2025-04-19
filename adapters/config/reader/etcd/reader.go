package etcd

import (
	"WgInspector/adapters/config/reader"
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"WgInspector/utils"
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/18
 */

func init() {
	config2.RegisterReader("etcd", &ConfigReaderEtcd{})
}

type ConfigReaderEtcd struct {
	client *clientv3.Client
	ctx    context.Context
	cfg    config.ConfigIndex
	parser config.Parser
}

var _ config.Reader = (*ConfigReaderEtcd)(nil)

func (*ConfigReaderEtcd) NewReader(option utils.Option) (_ config.Reader, err error) {
	//todo: initConfig read option
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{option.GetOrDefault("etcd_endpoint", "http://localhost:2379")},
		DialTimeout: 5 * time.Second,
		Username:    option.GetOrDefault("username", ""),
		Password:    option.GetOrDefault("password", ""),
	})
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("config reader - etcd: etcd client init fail, client is nil")
	}
	parser, err := config2.GetParser(option.GetOrDefault(reader.OptionParser, "json"))
	if err != nil {
		return nil, err
	}
	return &ConfigReaderEtcd{
		client: client,
		ctx:    context.TODO(),
		parser: parser,
	}, err
}

func (c *ConfigReaderEtcd) ReadConfig() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("config reader - etcd: Read Config fail: %v", r)
		}
	}()
	var meta config.ConfigMeta
	meta.DBs = readAndParse[config.DBConfig](c)
	meta.AgentTasks = readAndParse[config.AgentTaskConfig](c)
	meta.KnowledgeBases = readAndParse[config.KnowledgeBaseConfig](c)
	meta.Tasks = readAndParse[config.TaskConfig](c)
	meta.Logs = readAndParse[config.LogConfig](c)
	meta.Alerts = readAndParse[config.AlertConfig](c)
	if agentConfig := readAndParse[config.AgentConfig](c); len(agentConfig) > 0 {
		meta.Agent = agentConfig[0]
	}
	if inspConfig := readAndParse[*config.InspTree](c); len(inspConfig) > 0 {
		meta.Insp = inspConfig[0]
	}
	return config2.SetConfigMeta(meta)
}

func readAndParse[T config.ConfigType](c *ConfigReaderEtcd) []T {
	//从etcd中获取配置
	var t T
	name, err := config.GetConfigTypeName(t)
	if err != nil {
		return nil
	}
	key := "config/" + name
	ctx, _ := context.WithTimeout(c.ctx, time.Second*3)
	resp, err := c.client.Get(ctx, key)
	if err != nil || len(resp.Kvs) == 0 {
		return nil
	}
	data := resp.Kvs[0].Value

	f := func(res any, err error) []T {
		if err != nil {
			return nil
		}
		return res.([]T)
	}

	//解析配置
	switch any(t).(type) {
	case config.DBConfig:
		group, err := c.parser.ParseConfig(data)
		return f(group.DBs, err)
	case config.LogConfig:
		group, err := c.parser.ParseConfig(data)
		return f(group.Logs, err)
	case config.AlertConfig:
		group, err := c.parser.ParseConfig(data)
		return f(group.Alerts, err)
	case config.AgentTaskConfig:
		group, err := c.parser.ParseAgent(data)
		return f(group.AgentTasks, err)
	case config.AgentConfig:
		group, err := c.parser.ParseAgent(data)
		return f([]T{any(group.Agent).(T)}, err)
	case config.KnowledgeBaseConfig:
		group, err := c.parser.ParseAgent(data)
		return f(group.KnowledgeBases, err)
	case config.InspTree, *config.InspTree:
		insp, err := c.parser.ParseInspector(data)
		return f([]T{any(insp).(T)}, err)
	case config.TaskConfig:
		task, err := c.parser.ParseTask(data)
		return f(task.Tasks, err)
	}
	return nil
}

func (c *ConfigReaderEtcd) SaveConfig(data config.Id) error {
	config2.RLock()
	defer config2.RUnlock()
	configTypeName, err := config.GetConfigTypeName(data)
	if err != nil {
		return err
	}
	//1. 从meta中取出type对应配置
	switch configTypeName {
	case config.TypeDB:
		return c.encodeAndPut(configTypeName, config2.Meta.DBs)
	case config.TypeLog:
		return c.encodeAndPut(configTypeName, config2.Meta.Logs)
	case config.TypeAlert:
		return c.encodeAndPut(configTypeName, config2.Meta.Alerts)
	case config.TypeAgentTask:
		return c.encodeAndPut(configTypeName, config2.Meta.AgentTasks)
	case config.TypeInspector:
		return c.encodeAndPut(configTypeName, config2.Meta.Insp)
	case config.TypeKBase:
		return c.encodeAndPut(configTypeName, config2.Meta.KnowledgeBases)
	case config.TypeTask:
		return c.encodeAndPut(configTypeName, config2.Meta.Tasks)
	case config.TypeAgent:
		return c.encodeAndPut(configTypeName, config2.Meta.Agent)
	default:
		return fmt.Errorf("config read - etcd: config type name not suppose: %s\n", configTypeName)
	}
}

func (c *ConfigReaderEtcd) encodeAndPut(typeName string, val any) error {
	encoded, err := c.parser.Encode(val)
	if err != nil {
		return err
	}
	_, err = c.client.Put(c.ctx, "config/"+typeName, string(encoded))
	return err
}

func (c *ConfigReaderEtcd) Watch() {
	//1. 监听前缀
	watchChan := c.client.Watch(c.ctx, "config/", clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			if len(event.Kv.Key) <= 7 {
				continue
			}
			key := string(event.Kv.Key)[7:]
			var err error
			switch key {
			case config.TypeLog:
				err = config2.AppendConfigs(readAndParse[config.LogConfig](c)...)
			case config.TypeDB:
				err = config2.AppendConfigs(readAndParse[config.DBConfig](c)...)
			case config.TypeAlert:
				err = config2.AppendConfigs(readAndParse[config.AlertConfig](c)...)
			case config.TypeTask:
				err = config2.AppendConfigs(readAndParse[config.TaskConfig](c)...)
			case config.TypeAgent:
				err = config2.AppendConfigs(readAndParse[config.AgentConfig](c)...)
			case config.TypeAgentTask:
				err = config2.AppendConfigs(readAndParse[config.AgentTaskConfig](c)...)
			case config.TypeKBase:
				err = config2.AppendConfigs(readAndParse[config.KnowledgeBaseConfig](c)...)
			case config.TypeInspector:
				err = config2.AppendConfigs(readAndParse[config.InspTree](c)...)
			default:
				err = fmt.Errorf("key not suppose: %s", string(event.Kv.Key))
			}
			if err != nil {
				log.Printf("config reader - etcd: watch config fail: %s\n", err)
			}
		}
	}
}
