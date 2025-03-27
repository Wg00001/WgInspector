package etcd

import (
	"WgInspector/adapters/config/reader"
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"WgInspector/utils"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
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
	meta   config.ConfigMeta
}

var _ config.Reader = (*ConfigReaderEtcd)(nil)

func (c ConfigReaderEtcd) NewReader(option utils.Option) (_ config.Reader, err error) {
	c.client, err = clientv3.New(clientv3.Config{
		Username: option.GetOrDefault("username", ""),
		Password: option.GetOrDefault("password", ""),
	})
	c.parser, err = config2.GetParser(option.GetOrDefault(reader.OptionParser, "json"))
	if err != nil {
		return nil, err
	}
	return c, err
}

func (c ConfigReaderEtcd) ReadConfig() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("config reader - etcd: Read Config fail: %v", r)
		}
	}()
	c.meta.DBs = readAndParse[config.DBConfig](&c)
	c.meta.AgentTasks = readAndParse[config.AgentTaskConfig](&c)
	c.meta.KnowledgeBases = readAndParse[config.KnowledgeBaseConfig](&c)
	c.meta.Tasks = readAndParse[config.TaskConfig](&c)
	c.meta.Logs = readAndParse[config.LogConfig](&c)
	c.meta.Alerts = readAndParse[config.AlertConfig](&c)
	c.meta.Agent = readAndParse[config.AgentConfig](&c)[0]
	c.meta.Insp = readAndParse[*config.InspTree](&c)[0]
	return
}

func readAndParse[T config.ConfigType](c *ConfigReaderEtcd) []T {
	//从etcd中获取配置
	var t T
	name, err := config.GetConfigTypeName(t)
	if err != nil {
		return nil
	}
	key := "config/" + name
	resp, err := c.client.Get(c.ctx, key)
	if err != nil || len(resp.Kvs) == 0 {
		return nil
	}
	data := resp.Kvs[0].Value

	//解析配置
	switch any(t).(type) {
	case config.DBConfig:
		group, _ := c.parser.ParseConfig(data)
		return any(group.DBs).([]T)
	case config.LogConfig:
		group, _ := c.parser.ParseConfig(data)
		return any(group.Logs).([]T)
	case config.AlertConfig:
		group, _ := c.parser.ParseConfig(data)
		return any(group.Alerts).([]T)
	case config.AgentTaskConfig:
		group, _ := c.parser.ParseAgent(data)
		return any(group.AgentTasks).([]T)
	case config.AgentConfig:
		group, _ := c.parser.ParseAgent(data)
		return []T{any(group.Agent).(T)}
	case config.KnowledgeBaseConfig:
		group, _ := c.parser.ParseAgent(data)
		return any(group.KnowledgeBases).([]T)
	case config.InspTree, *config.InspTree:
		insp, _ := c.parser.ParseInspector(data)
		//todo: nil panic test
		return []T{any(insp).(T)}
	case config.TaskConfig:
		task, _ := c.parser.ParseTask(data)
		return any(task.Tasks).([]T)
	}
	return nil
}

func (c ConfigReaderEtcd) SaveIntoConfig() {
	config2.SetConfigMeta(c.meta)
}

func (c ConfigReaderEtcd) Watch() {
	//TODO implement me
	panic("implement me")
}
