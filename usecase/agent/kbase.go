package agent

import (
	"WgInspector/entities/agent"
	"WgInspector/usecase/agent/kbase"
	"encoding/json"
)

/**
 * @description: TODO
 * @author Wg
 * @date 2025/4/5
 */

// KBaseSave 保存到知识库中。
// 用户收到消息后，对知识库入库消息进行确认，对于确认入库的消息使用此函数入库
func KBaseSave(msg string) error {
	//todo:置信度评估
	//todo: 对比去重

	//人工审核已通过
	//1. 关键词提取
	var report AnalysisReport
	err := json.Unmarshal([]byte(msg), &report)
	if err != nil {
		return err
	}

	//2. 格式转换
	docs := make([]agent.Document, 0, len(report.InspectLogAnalyze))
	for i := range report.InspectLogAnalyze {
		docs = append(docs, agent.Document{
			Content: report.InspectLogAnalyze[i].String(),
			Metadata: map[string]interface{}{
				"metrics": report.InspectLogAnalyze[i].Metrics,
				"belongs": report.InspectLogAnalyze[i].Belongs,
			},
		})
	}

	//3. 持久化到知识库中
	return kbase.Save(docs)
}
