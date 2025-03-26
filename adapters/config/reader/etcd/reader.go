package etcd

import (
	"WgInspector/adapters/config/parser/yaml"
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"context"
	"github.com/coreos/etcd/clientv3"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/18
 */

type ConfigReaderEtcd struct {
	client      *clientv3.Client
	rootContext context.Context
	cfg         config.ConfigIndex
	parser      yaml.ConfigYamlParser
	meta        config.ConfigMeta
}

var _ config.Reader = (*ConfigReaderEtcd)(nil)

func (c ConfigReaderEtcd) NewReader(option map[string]string) (_ config.Reader, err error) {
	c.client, err = clientv3.New(clientv3.Config{})
	return c, err
}

func (c ConfigReaderEtcd) ReadConfig() error {
	_, err := c.client.Get(c.rootContext, "config")
	if err != nil {
		return err
	}
	return nil
}

func (c ConfigReaderEtcd) ReadInspector() error {
	//TODO implement me
	panic("implement me")
}

func (c ConfigReaderEtcd) ReadAgent() error {
	//TODO implement me
	panic("implement me")
}

func (c ConfigReaderEtcd) SaveIntoConfig() {
	config2.SetConfigMeta(c.meta)
}

func (c ConfigReaderEtcd) Watch() {
	//TODO implement me
	panic("implement me")
}
