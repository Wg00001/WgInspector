package _default

import (
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	ai2 "WgInspector/usecase/agent/analyzer"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/1
 */

func init() {
	ai2.RegisterDriver("default", AnalyzerDefault{})
	//ai2.RegisterDriver("", AnalyzerDefault{})
}

type AnalyzerDefault struct {
}

func (a AnalyzerDefault) Init(*config.AgentConfig) (agent.Analyzer, error) {
	return AnalyzerDefault{}, nil
}

func (a AnalyzerDefault) Analyze(*agent.AnalyzeContent) (string, error) {
	return "analyzer didn't define", nil
}

var _ agent.Analyzer = (*AnalyzerDefault)(nil)
