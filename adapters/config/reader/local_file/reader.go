package local_file

import (
	"WgInspector/adapters/config/reader"
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	"WgInspector/utils"
	"fmt"
	"os"
	"strings"
)

/**
 * @description: insp
 * @author Wg
 * @date 2025/1/19
 */

const (
	optionFilepath   = "filepath"
	optionConfigName = "config"
	optionInspName   = "inspect"
	optionAgentName  = "agent"
	optionTaskName   = "task"
)

func init() {
	config2.RegisterReader("file", &ConfigReaderLocalFile{})
	config2.RegisterReader("local_file", &ConfigReaderLocalFile{})
}

type ConfigReaderLocalFile struct {
	FilePath   string
	ConfigName string
	InspName   string
	AgentName  string
	TaskName   string
	parser     config.Parser
	meta       config.ConfigMeta
}

var _ config.Reader = (*ConfigReaderLocalFile)(nil)

func (c *ConfigReaderLocalFile) NewReader(option utils.Option) (_ config.Reader, err error) {
	filepath, ok := option[optionFilepath]
	if !ok {
		return nil, fmt.Errorf("config reader: option deficiency - %s\n", optionFilepath)
	}
	configName := option.GetOrDefault(optionConfigName, optionConfigName+".yaml")
	inspName := option.GetOrDefault(optionInspName, optionInspName+".yaml")
	agentName, ok := option[optionAgentName]
	if !ok {
		agentName = optionAgentName + ".yaml"
	}
	parserDriver, ok := option[reader.OptionParser]
	if !ok {
		parserDriver = "yaml"
	}
	taskName, ok := option[optionTaskName]
	if !ok {
		taskName = optionTaskName + ".yaml"
	}
	if !strings.HasSuffix(filepath, "/") {
		filepath += "/"
	}
	parser, err := config2.GetParser(parserDriver)
	if err != nil {
		return nil, err
	}
	return &ConfigReaderLocalFile{
		FilePath:   filepath,
		ConfigName: configName,
		InspName:   inspName,
		AgentName:  agentName,
		parser:     parser,
		TaskName:   taskName,
	}, nil
}

func (c *ConfigReaderLocalFile) ReadConfig() (err error) {
	file, err := os.ReadFile(c.FilePath + c.ConfigName)
	if err != nil {
		return err
	}
	c.meta.CommonConfigGroup, err = c.parser.ParseConfig(file)

	file, err = os.ReadFile(c.FilePath + c.InspName)
	if err != nil {
		return err
	}
	c.meta.Insp, err = c.parser.ParseInspector(file)

	file, err = os.ReadFile(c.FilePath + c.TaskName)
	if err != nil {
		return err
	}
	c.meta.TaskConfigGroup, err = c.parser.ParseTask(file)

	file, err = os.ReadFile(c.FilePath + c.AgentName)
	if err != nil {
		return err
	}
	c.meta.AgentConfigGroup, err = c.parser.ParseAgent(file)
	return config2.SetConfigMeta(c.meta)
}

func (c *ConfigReaderLocalFile) SaveConfig(string) error {
	return nil
}

func (c *ConfigReaderLocalFile) Watch() {
	//TODO implement me
	panic("implement me")
}
