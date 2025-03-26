package openai

import (
	"WgInspector/entities/agent"
	"WgInspector/entities/config"
	ai2 "WgInspector/usecase/agent/analyzer"
	"WgInspector/utils"
	"context"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"log"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/3/5
 */

func init() {
	ai2.RegisterDriver("openai", AnalyzerAgent{})
}

type AnalyzerAgent config.AgentConfig

var _ agent.Analyzer = (*AnalyzerAgent)(nil)

func (a AnalyzerAgent) Init(aiConfig *config.AgentConfig) (agent.Analyzer, error) {
	return AnalyzerAgent(*aiConfig), nil
}

func (a AnalyzerAgent) Analyze(content *agent.AnalyzeContent) (string, error) {
	llm, err := openai.New(openai.WithBaseURL(a.Url), openai.WithToken(a.ApiKey), openai.WithModel(a.Model))
	if err != nil {
		return "", err
	}

	if content.SystemMsg == "" {
		if a.SystemMessage == "" {
			content.SystemMsg = utils.DefaultSystemMessage
		} else {
			content.SystemMsg = a.SystemMessage
		}
	}

	msg := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem,
			content.SystemMsg,
			func() string {
				if content.KBaseMsg != "" {
					return "\n\n可以参考以下知识库内容进行分析:\n" + content.KBaseMsg
				}
				return ""
			}(),
		),
		llms.TextParts(llms.ChatMessageTypeHuman,
			"\n\n日志内容:\n"+content.UserMsg),
	}
	log.Printf("发送消息: %v\n", msg)

	ctx := context.Background()
	resp, err := llm.GenerateContent(ctx, msg)
	if err != nil {
		return "", err
	}
	log.Printf("收到消息：%v\n", resp.Choices[0].Content)

	return resp.Choices[0].Content, nil
}
