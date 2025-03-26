package agent

import "WgInspector/entities/config"

/**
 * @description: Ai的itf
 * @author Wg
 * @date 2025/2/22
 */

//Ai模块与logger模块、alert模块高度耦合。用例需要从logger中读取日志，然后发送给Ai读取，再由Alert发送给用户。
//包括以下步骤：
//1. 配置文件读取AiTaskConfig，指定目标LogID以及Filter，由Cron设置定时任务。
//2. 发送给Ai分析并获得分析结果的功能由本接口实现。使用usecase/openai/format来标准化发送和接受。
//3. 将分析结果发送给Alert

// Analyzer 无状态，接受Temporary并发送给ai，然后解析并返回结果
type Analyzer interface {
	Init(*config.AgentConfig) (Analyzer, error)
	// Analyze 将可供ai分析的string发送给AI进行分析，并解析返回结果。content使用指针传递，避免过大
	Analyze(content *AnalyzeContent) (string, error)
}

type AnalyzeContent struct {
	SystemMsg string //提示词
	UserMsg   string //日志数据
	KBaseMsg  string //从知识库中获取的相关知识
}
