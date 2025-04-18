package yaml

import (
	"WgInspector/entities/config"
	config2 "WgInspector/usecase/config"
	insp2 "WgInspector/usecase/insp"
	"WgInspector/utils"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/18
 */

func init() {
	config2.RegisterParser("yaml", &ConfigYamlParser{})
}

type ConfigYamlParser struct {
}

type ConfigYaml struct {
	Default           config.InitConfig        `yaml:"default"`
	DBConfigs         []config.DBConfig        `yaml:"db"`
	LogConfigOrigin   []map[string]interface{} `yaml:"log"`
	LogConfig         []config.LogConfig       `yaml:"-"`
	AlertConfigOrigin []map[string]interface{} `yaml:"alert"`
	AlertConfig       []config.AlertConfig     `yaml:"-"`
}

type TaskYaml struct {
	TaskConfigs []config.TaskConfig `yaml:"task"`
}

type AgentConfigYaml struct {
	AiConfig           config.AgentConfig           `yaml:"agent"`
	AiTaskConfigOrigin []map[string]interface{}     `yaml:"agenttask"`
	AiTaskConfig       []config.AgentTaskConfig     `yaml:"-"`
	KBaseConfigOrigin  []map[string]interface{}     `yaml:"kbase"`
	KBaseConfig        []config.KnowledgeBaseConfig `yaml:"-"`
}

var _ config.Parser = (*ConfigYamlParser)(nil)

func (c *ConfigYamlParser) ParseConfig(file []byte) (_ config.CommonConfigGroup, err error) {
	cyaml := ConfigYaml{}
	file = []byte(strings.ToLower(string(file)))
	err = yaml.Unmarshal(file, &cyaml)
	if err != nil {
		return
	}
	//处理logger设置
	cyaml.LogConfig = make([]config.LogConfig, 0, len(cyaml.LogConfigOrigin))
	for _, o := range cyaml.LogConfigOrigin {
		cyaml.LogConfig = append(cyaml.LogConfig,
			func(origin map[string]interface{}) config.LogConfig {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("fail to read logger config")
					}
				}()
				cur := config.LogConfig{
					Identity: config.NewIdentity(origin["identity"]),
					Driver:   origin["driver"].(string),
					Option:   make(map[string]string),
				}
				for k, v := range origin {
					if str, ok := v.(string); ok {
						cur.Option[k] = str
					}
				}
				return cur
			}(o))
	}
	if err != nil {
		return
	}

	cyaml.AlertConfig = make([]config.AlertConfig, 0, len(cyaml.AlertConfigOrigin))
	for _, o := range cyaml.AlertConfigOrigin {
		cyaml.AlertConfig = append(cyaml.AlertConfig,
			func(origin map[string]interface{}) config.AlertConfig {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("fail to read alert config")
					}
				}()
				cur := config.AlertConfig{
					Identity: config.NewIdentity(origin["identity"]),
					Driver:   origin["driver"].(string),
					Option:   make(map[string]string),
				}
				for k, v := range origin {
					if str, ok := v.(string); ok {
						cur.Option[k] = str
					}
				}
				return cur
			}(o))
	}
	return config.CommonConfigGroup{
		DBs:    cyaml.DBConfigs,
		Logs:   cyaml.LogConfig,
		Alerts: cyaml.AlertConfig,
	}, nil
}

func (c *ConfigYamlParser) ParseTask(file []byte) (_ config.TaskConfigGroup, err error) {
	cyaml := TaskYaml{}
	file = []byte(strings.ToLower(string(file)))
	err = yaml.Unmarshal(file, &cyaml)
	if err != nil {
		return
	}
	return config.TaskConfigGroup{Tasks: cyaml.TaskConfigs}, err
}

func (c *ConfigYamlParser) ParseInspector(file []byte) (_ *config.InspTree, err error) {
	var origin map[string]interface{}
	err = yaml.Unmarshal(file, &origin)
	if err != nil {
		return
	}
	inspTree := config.NewTree()
	var dfs func(nowPath string, node map[string]interface{}) error
	dfs = func(nowPath string, node map[string]interface{}) error {
		for k, v := range node {
			switch t := v.(type) {
			case string: //配置里直接是string说明是SQL，未配置alert
				n, err := insp2.NodeBuilder{}.WithName(config.Identity(k)).WithSQL(t).Build()
				//n, err := insp2.NodeBuilder{}.WithName(k).WithSQL(t).WithEmptyAlert().Build()
				if err != nil {
					return err
				}
				err = inspTree.AddChild(nowPath, &n)
				if err != nil {
					return err
				}

			case map[string]interface{}: //是map说明有配置alert, 进行解析
				nb, err := ParseMap(insp2.NodeBuilder{}.WithName(config.Identity(k)), t)
				if err != nil {
					return err
				}
				n, err := nb.Build()
				if err != nil {
					return err

				}
				err = inspTree.AddChild(nowPath, &n)
				if err != nil {
					return err
				}
				err = dfs(nowPath+k, t)
				if err != nil {
					return err
				}

			}
		}
		return nil
	}
	return inspTree, dfs("", origin)
}

func (c *ConfigYamlParser) ParseAgent(file []byte) (_ config.AgentConfigGroup, err error) {
	file = []byte(strings.ToLower(string(file)))
	ayaml := AgentConfigYaml{}
	err = yaml.Unmarshal(file, &ayaml)
	if err != nil {
		return
	}

	// Process agent task configurations
	ayaml.AiTaskConfig = make([]config.AgentTaskConfig, 0, len(ayaml.AiTaskConfigOrigin))
	for _, origin := range ayaml.AiTaskConfigOrigin {
		funcResult := func(o map[string]interface{}) (result config.AgentTaskConfig) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("fail to read agent task config: %v", r)
				}
			}()

			m := utils.UseMap(o)

			// Basic configurations
			result.Identity = config.Identity(m.GetString("identity"))
			result.LogID = config.Identity(m.GetInt("logid"))
			result.AlertID = config.Identity(m.GetInt("alertid"))
			result.KBaseResults = m.GetInt("kbaseresults")
			result.KBaseMaxLen = m.GetInt("kbasemaxlen")
			result.SystemMessage = m.GetString("systemmessage")

			// Cron configuration
			if cronMap := m.GetMap("cron"); cronMap != nil {
				result.Cron = &config.Cron{
					CronTab: cronMap.GetString("crontab"),
					Duration: func() time.Duration {
						duration, err := time.ParseDuration(cronMap.GetString("duration"))
						if err != nil {
							return 0
						}
						return duration
					}(),
					AtTime:  parseStringSlice(cronMap["attime"]),
					Weekly:  parseWeekdaySlice(cronMap["weekly"]),
					Monthly: parseIntSlice(cronMap["monthly"]),
				}
			}

			// Log filter configuration
			if logFilterMap := m.GetMap("logfilter"); logFilterMap != nil {
				result.LogFilter = config.LogFilter{
					StartTime: parseTime(logFilterMap.GetString("starttime")),
					EndTime:   parseTime(logFilterMap.GetString("endtime")),
					TaskNames: parseNames(logFilterMap["tasknames"]),
					DBIDs:     parseNames(logFilterMap["dbnames"]),
					TaskIDs:   parseNames(logFilterMap["taskids"]),
					InspNames: parseNames(logFilterMap["inspnames"]),
				}
			}

			// Knowledge base references
			result.KBase = parseNames(m["kbase"])
			return
		}(origin)
		ayaml.AiTaskConfig = append(ayaml.AiTaskConfig, funcResult)
	}
	if err != nil {
		return
	}

	// Process knowledge base configurations
	ayaml.KBaseConfig = make([]config.KnowledgeBaseConfig, 0, len(ayaml.KBaseConfigOrigin))
	for _, origin := range ayaml.KBaseConfigOrigin {
		funcResult := func(o map[string]interface{}) (result config.KnowledgeBaseConfig) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("fail to read knowledge base config: %v", r)
				}
			}()

			m := utils.UseMap(o)
			result.Identity = config.Identity(m.GetString("name"))
			result.Driver = m.GetString("driver")
			result.Value = m
			return
		}(origin)
		ayaml.KBaseConfig = append(ayaml.KBaseConfig, funcResult)
	}
	return config.AgentConfigGroup{
		Agent:          ayaml.AiConfig,
		AgentTasks:     ayaml.AiTaskConfig,
		KnowledgeBases: ayaml.KBaseConfig,
	}, nil
}
func (c *ConfigYamlParser) Encode(arg any) ([]byte, error) {
	return nil, nil
}
